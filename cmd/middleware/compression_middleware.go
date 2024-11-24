package middleware

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/middleware"
)

type AdaptiveCompressionConfig struct {
	MinSize      int
	Level        int   // Compression level (1-9)
	MaxRetries   int
	RetryDelay   time.Duration

	// Adaptive compression config
	EnableAdaptive       bool
	UpdateInterval       time.Duration
	MaxLatency           time.Duration

	// Adjustment thresholds
	MinLevel             int  // Minimum compression level
	MaxLevel             int  // Maximum compression level
	SampleThreshold      int  // Min number of requests before adjusting
	TargetLatency        time.Duration // Target latency for compression
	AdjustmentInterval   time.Duration // How often to adjust compression level

	MinCompressionRatio  float64   // Minimum acceptable ratio
	MaxCompressionRatio  float64   // Maximum acceptable ratio
	TargetCompressionRatio float64 // Ideal compression ratio
}

type compressionResponseWriter struct {
	http.ResponseWriter
	mu             sync.Mutex
	written        int64
	originalSize   int64
	path           string
	monitor        *CompressionMonitor
	startTime      time.Time
	err            error
}

var DefaultCompressionConfig = AdaptiveCompressionConfig{
	MinSize:       1024,     // 1KB
	Level  :       5,        // Default medium compression
	MaxRetries:    3,
	RetryDelay:    100 * time.Millisecond,

	// Adaptive compression config
	EnableAdaptive:           true,
	UpdateInterval:           1 * time.Hour,
	MinCompressionRatio:      0.3,    // Minimum 30% compression
	MaxCompressionRatio:      0.9,    // Maximum 90% compression
	TargetCompressionRatio:   0.5,  // Target 50% compression
	MaxLatency:               50 * time.Millisecond,

	// Adjustment thresholds
	MinLevel:               1,
	MaxLevel:               9,
	SampleThreshold:        10,
	TargetLatency:          20 * time.Millisecond,
	AdjustmentInterval:     1 * time.Minute,

}

func NewAdaptiveCompression(monitor *CompressionMonitor) func(http.Handler) http.Handler {
	// Guard clauses
	if monitor == nil {
		panic("compression monitor cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}

	// Attempt to load config from env
	cfg, err := LoadCompressionConfig()
	if err != nil {
		// Use default config if loading fails
		cfg = &DefaultCompressionConfig
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid compression config: %v", err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Error recovery
			defer func() {
				if err := recover(); err != nil {
					// Log panic
					stack := debug.Stack()
					log.Printf("Compression panic recovered: %v\nStack: %s", err, stack)

					// Increment failure count in stats
					if monitor != nil {
						monitor.recordFailure(r.URL.Path, fmt.Sprintf("panic: %v", err))
					}

					// Fallback to uncompressed response
					next.ServeHTTP(w, r)
				}
			}()

			// Skip compression for non-compressible requests
			if !shouldCompress(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Create safe response writer
			crw := &compressionResponseWriter{
				ResponseWriter: w,
				path:          r.URL.Path,
				monitor:       monitor,
				startTime:     time.Now(),
			}

			// Get compression config and apply
			pathConfig := monitor.getPathConfig(r.URL.Path)
			compressor := middleware.Compress(pathConfig.Level)
			handler := compressor(next)
			handler.ServeHTTP(crw, r)

			// Check for compression errors
			if crw.err != nil {
				monitor.recordFailure(r.URL.Path, crw.err.Error())
				next.ServeHTTP(w, r)
				return
			}

			// Update stats after response
			monitor.updateStats(
				r.URL.Path,
				crw.written,             // compressed size
				crw.originalSize,        // original size
				time.Since(crw.startTime))

			if cfg.EnableAdaptive {
				monitor.adjustCompressionLevel(r.URL.Path, *cfg)
			}
		})
	}
}

// Add this method to AdaptiveCompressionConfig
func (cfg *AdaptiveCompressionConfig) Validate() error {
	// Basic validation
	if cfg.MinSize < 0 {
			return fmt.Errorf("minSize must be non-negative, got %d", cfg.MinSize)
	}
	if cfg.Level < 1 || cfg.Level > 9 {
			return fmt.Errorf("level must be between 1-9, got %d", cfg.Level)
	}
	if cfg.MaxRetries < 0 {
			return fmt.Errorf("maxRetries must be non-negative, got %d", cfg.MaxRetries)
	}

	// Time-based validation
	if cfg.RetryDelay < time.Millisecond {
			return fmt.Errorf("retryDelay must be at least 1ms, got %v", cfg.RetryDelay)
	}
	if cfg.UpdateInterval < time.Second {
			return fmt.Errorf("updateInterval must be at least 1s, got %v", cfg.UpdateInterval)
	}

	// Compression ratio validation
	if cfg.MinCompressionRatio < 0 || cfg.MinCompressionRatio > 1 {
			return fmt.Errorf("minCompressionRatio must be between 0-1, got %f", cfg.MinCompressionRatio)
	}
	if cfg.MaxCompressionRatio < cfg.MinCompressionRatio || cfg.MaxCompressionRatio > 1 {
			return fmt.Errorf("maxCompressionRatio must be between MinCompressionRatio and 1, got %f", cfg.MaxCompressionRatio)
	}
	if cfg.TargetCompressionRatio < cfg.MinCompressionRatio || cfg.TargetCompressionRatio > cfg.MaxCompressionRatio {
			return fmt.Errorf("targetCompressionRatio must be between min and max ratios, got %f", cfg.TargetCompressionRatio)
	}

	return nil
}


// Write implements the http.ResponseWriter interface
func (crw *compressionResponseWriter) Write(b []byte) (int, error) {
	crw.mu.Lock()
	defer crw.mu.Unlock()

	if crw.err != nil {
		return 0, crw.err
	}

	// Store original size before compression
	crw.originalSize = int64(len(b))

	n, err := crw.ResponseWriter.Write(b)

	// Compressed size
	crw.written += int64(n)
	crw.err = err
	return n, err
}

// Helpers to determine if a request should be compressed
func isCompressible(contentType string) bool {
	compressibleTypes := []string{
		"text/",
		"application/json",
	}

	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}

func shouldCompress(r *http.Request) bool {
	// Check Accept-Encoding header
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}

	// Check Content-Type header if available
	contentType := r.Header.Get("Content-Type")
	return contentType == "" || isCompressible(contentType)
}

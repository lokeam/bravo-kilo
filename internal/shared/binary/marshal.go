package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
)

const MaxMemoryLimit = 10 * 1024 * 1024 // 10MB limit

// Reusable function to marshal any data into a binary format
func MarshalBinary(v any) ([]byte, error) {
	// Setup contextual logger
	logger := slog.With(
		"operation", "binary_marshal",
		"dataType", fmt.Sprintf("%T", v),
)

	// Guard clause for nil input
	if v == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

// First convert the struct to JSON
jsonData, err := json.Marshal(v)
if err != nil {
	logger.Error("json marshal failed",
			"error", err,
			"dataType", fmt.Sprintf("%T", v),
	)
	return nil, fmt.Errorf("json marshal failed: %w", err)
}

// Check memory limit before proceeding
if len(jsonData) > MaxMemoryLimit {
	logger.Error("data exceeds memory limit",
			"size", len(jsonData),
			"limit", MaxMemoryLimit,
	)
	return nil, fmt.Errorf("data size %d exceeds memory limit %d", len(jsonData), MaxMemoryLimit)
}

// Create a buffer with capacity hint to avoid reallocations
buf := bytes.NewBuffer(make([]byte, 0, len(jsonData)+4)) // +4 for length prefix

// Write length as uint32 (4 bytes)
if err := binary.Write(buf, binary.LittleEndian, uint32(len(jsonData))); err != nil {
	logger.Error("failed to write length prefix",
			"error", err,
			"dataLength", len(jsonData),
	)
	return nil, fmt.Errorf("failed to write length prefix: %w", err)
}

// Write JSON data
if _, err := buf.Write(jsonData); err != nil {
	logger.Error("failed to write json data",
			"error", err,
			"bufferSize", buf.Len(),
			"jsonLength", len(jsonData),
	)
	return nil, fmt.Errorf("failed to write json data: %w", err)
}

logger.Debug("binary marshaling completed",
		"totalSize", buf.Len(),
		"jsonSize", len(jsonData),
)

return buf.Bytes(), nil
}
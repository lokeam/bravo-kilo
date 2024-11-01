package redis

import (
	"time"
)

func NewTimeWindow(duration time.Duration) *TimeWindow {
    return &TimeWindow{
        duration: duration,
        samples:  make([]Sample, 0),
    }
}

func (w *TimeWindow) Add(value float64) {
    w.mu.Lock()
    defer w.mu.Unlock()

    now := time.Now()
    w.samples = append(w.samples, Sample{
        timestamp: now,
        value:     value,
    })

    // Clean old samples
    w.cleanup(now)
}

func (w *TimeWindow) GetRate() float64 {
    w.mu.RLock()
    defer w.mu.RUnlock()

    if len(w.samples) == 0 {
        return 0
    }

    var sum float64
    for _, sample := range w.samples {
        sum += sample.value
    }

    return sum / float64(len(w.samples))
}

func (w *TimeWindow) cleanup(now time.Time) {
    cutoff := now.Add(-w.duration)
    newSamples := make([]Sample, 0, len(w.samples))

    for _, sample := range w.samples {
        if sample.timestamp.After(cutoff) {
            newSamples = append(newSamples, sample)
        }
    }

    w.samples = newSamples
}
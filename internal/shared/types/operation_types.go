package types

import (
	"time"
)

type Operation struct {
	Name      string
	StartTime time.Time
	Budget    float64
}
package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	L := &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
	}
	if clock != nil {
		L.last = clock.Now()
	}
	if L.rate <= 0 {
		L.rate = 0
	}

	return L
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.clock == nil {
		return false
	}
	if n > l.burst || n <= 0 {
		return false
	}

	if l.clock.Now().Sub(l.last) > 0 {
		l.tokens += l.rate * time.Duration.Seconds(l.clock.Now().Sub(l.last))
		l.last = l.clock.Now()
	}
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	if l.tokens < float64(n) {
		return false
	}

	l.tokens -= float64(n)
	return true
}

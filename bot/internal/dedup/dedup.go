package dedup

import (
	"sync"
	"time"
)

type Deduper struct {
	window time.Duration

	mu            sync.Mutex
	last          map[string]time.Time
	lastPrune     time.Time
	pruneInterval time.Duration
}

func New(window time.Duration) *Deduper {
	pruneEvery := time.Minute
	if window > 0 && window < pruneEvery {
		pruneEvery = window / 2
		if pruneEvery < time.Second {
			pruneEvery = time.Second
		}
	}

	return &Deduper{
		window:        window,
		last:          make(map[string]time.Time),
		pruneInterval: pruneEvery,
	}
}

// Allow: true — можно отправлять алерт; false — дубликат внутри окна window.
func (d *Deduper) Allow(key string, now time.Time) bool {
	if key == "" {
		return true
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.last[key]; ok && now.Sub(t) < d.window {
		d.maybePruneLocked(now)
		return false
	}

	d.last[key] = now
	d.maybePruneLocked(now)
	return true
}

func (d *Deduper) maybePruneLocked(now time.Time) {
	if d.pruneInterval <= 0 {
		return
	}
	if !d.lastPrune.IsZero() && now.Sub(d.lastPrune) < d.pruneInterval {
		return
	}
	d.pruneLocked(now)
	d.lastPrune = now
}

func (d *Deduper) pruneLocked(now time.Time) {
	for k, t := range d.last {
		if now.Sub(t) >= d.window {
			delete(d.last, k)
		}
	}
}

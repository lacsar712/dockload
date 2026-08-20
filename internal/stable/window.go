// Package stable implements a sliding window stability detector for raw counts.
package stable

import (
	"math"
)

// Window retains the most recent raw sensor samples and evaluates stability.
type Window struct {
	size int
	eps  float64
	buf  []int64
	head int
	count int
}

// NewWindow creates a fixed-size stability window.
func NewWindow(size int, eps float64) *Window {
	if size < 1 {
		size = 1
	}
	return &Window{
		size: size,
		eps:  eps,
		buf:  make([]int64, size),
	}
}

// Reset clears all samples without reallocation.
func (w *Window) Reset() {
	w.head = 0
	w.count = 0
	for i := range w.buf {
		w.buf[i] = 0
	}
}

// Push adds a sample, evicting the oldest when full.
func (w *Window) Push(raw int64) {
	w.buf[w.head] = raw
	w.head = (w.head + 1) % w.size
	if w.count < w.size {
		w.count++
	}
}

// Count returns number of samples currently held.
func (w *Window) Count() int {
	return w.count
}

// Full reports whether the window has N samples.
func (w *Window) Full() bool {
	return w.count == w.size
}

// Samples returns a copy of active samples in chronological order.
func (w *Window) Samples() []int64 {
	out := make([]int64, w.count)
	if w.count == 0 {
		return out
	}
	start := 0
	if w.count == w.size {
		start = w.head
	}
	for i := 0; i < w.count; i++ {
		idx := (start + i) % w.size
		out[i] = w.buf[idx]
	}
	return out
}

// Mean returns arithmetic mean of active samples.
func (w *Window) Mean() float64 {
	if w.count == 0 {
		return 0
	}
	var sum float64
	start := 0
	if w.count == w.size {
		start = w.head
	}
	for i := 0; i < w.count; i++ {
		idx := (start + i) % w.size
		sum += float64(w.buf[idx])
	}
	return sum / float64(w.count)
}

// StdDev returns population standard deviation of active samples.
func (w *Window) StdDev() float64 {
	if w.count == 0 {
		return 0
	}
	mean := w.Mean()
	var sumSq float64
	start := 0
	if w.count == w.size {
		start = w.head
	}
	for i := 0; i < w.count; i++ {
		idx := (start + i) % w.size
		diff := float64(w.buf[idx]) - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(w.count))
}

// Stable reports whether window is full and stddev <= eps.
func (w *Window) Stable() bool {
	return w.Full() && w.StdDev() <= w.eps
}

// Eps returns the stability threshold.
func (w *Window) Eps() float64 {
	return w.eps
}

// Size returns configured window capacity.
func (w *Window) Size() int {
	return w.size
}

// Last returns the most recently pushed sample and whether one exists.
func (w *Window) Last() (int64, bool) {
	if w.count == 0 {
		return 0, false
	}
	idx := w.head - 1
	if idx < 0 {
		idx = w.size - 1
	}
	return w.buf[idx], true
}

package monitoring

import (
	"math/rand"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// RequestSampler determines whether a request should be sampled for monitoring
type RequestSampler interface {
	// ShouldSample determines if a request should be sampled
	ShouldSample(method string) bool

	// GetSamplingRate returns the current sampling rate
	GetSamplingRate() float64

	// UpdateSamplingRate updates the sampling rate
	UpdateSamplingRate(rate float64)

	// GetStats returns sampling statistics
	GetStats() *SamplingStats
}

// SamplingStats represents sampling statistics
type SamplingStats struct {
	TotalRequests int64   `json:"total_requests"`
	SampledRequests int64 `json:"sampled_requests"`
	SamplingRate float64 `json:"sampling_rate"`
	LastUpdated time.Time `json:"last_updated"`
}

// DefaultRequestSampler represents a default implementation of RequestSampler
type DefaultRequestSampler struct {
	enabled    bool
	rate       float64
	rng        *rand.Rand
	mu         sync.RWMutex
	stats      *SamplingStats
}

// RandomSampler uses random sampling
type RandomSampler struct {
	rate float64
	rng  *rand.Rand
}

// DeterministicSampler uses deterministic sampling based on method
type DeterministicSampler struct {
	rate       float64
	methods    map[string]bool
	counter    int64
	sampleStep int64
}

// NewRequestSampler creates a new request sampler
func NewRequestSampler(enabled bool, rate float64, log *logger.Logger) RequestSampler {
	if !enabled || rate <= 0 {
		return NewNoopRequestSampler()
	}

	return &DefaultRequestSampler{
		enabled: enabled,
		rate:    rate,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		stats: &SamplingStats{
			SamplingRate: rate,
			LastUpdated:  time.Now(),
		},
	}
}

// ShouldSample determines if a request should be sampled
func (s *DefaultRequestSampler) ShouldSample(method string) bool {
	if !s.enabled {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalRequests++

	// Random sampling based on configured rate
	shouldSample := s.rng.Float64() < s.rate

	if shouldSample {
		s.stats.SampledRequests++
	}

	return shouldSample
}

// GetSamplingRate returns the current sampling rate
func (s *DefaultRequestSampler) GetSamplingRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rate
}

// UpdateSamplingRate updates the sampling rate
func (s *DefaultRequestSampler) UpdateSamplingRate(rate float64) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rate = rate
	s.stats.SamplingRate = rate
	s.stats.LastUpdated = time.Now()
}

// GetStats returns sampling statistics
func (s *DefaultRequestSampler) GetStats() *SamplingStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid concurrent modifications
	return &SamplingStats{
		TotalRequests:   s.stats.TotalRequests,
		SampledRequests: s.stats.SampledRequests,
		SamplingRate:    s.stats.SamplingRate,
		LastUpdated:     s.stats.LastUpdated,
	}
}

// NewRandomSampler creates a new random sampler
func NewRandomSampler(rate float64) *RandomSampler {
	return &RandomSampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample implements random sampling
func (r *RandomSampler) ShouldSample(method string) bool {
	return r.rng.Float64() < r.rate
}

// NewDeterministicSampler creates a new deterministic sampler
func NewDeterministicSampler(rate float64) *DeterministicSampler {
	sampleStep := int64(1.0 / rate)
	if sampleStep < 1 {
		sampleStep = 1
	}

	return &DeterministicSampler{
		rate:       rate,
		methods:    make(map[string]bool),
		sampleStep: sampleStep,
	}
}

// ShouldSample implements deterministic sampling
func (d *DeterministicSampler) ShouldSample(method string) bool {
	d.counter++

	// Sample every N-th request
	shouldSample := d.counter%d.sampleStep == 0

	// Cache result for this method to ensure consistent sampling
	d.methods[method] = shouldSample

	return shouldSample
}

// AdaptiveSampler adjusts sampling rate based on request volume
type AdaptiveSampler struct {
	baseRate         float64
	maxRate          float64
	minRate          float64
	requestWindow    int64
	windowDuration   time.Duration
	requestCounts    []int64
	currentWindow    int
	lastWindowReset  time.Time
	rng              *rand.Rand
	mu               sync.RWMutex
}

// NewAdaptiveSampler creates a new adaptive sampler
func NewAdaptiveSampler(baseRate, maxRate, minRate float64, windowSize int64, windowDuration time.Duration) *AdaptiveSampler {
	return &AdaptiveSampler{
		baseRate:        baseRate,
		maxRate:         maxRate,
		minRate:         minRate,
		requestWindow:   windowSize,
		windowDuration:  windowDuration,
		requestCounts:   make([]int64, 100), // 100 windows max
		currentWindow:   0,
		lastWindowReset: time.Now(),
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample implements adaptive sampling
func (a *AdaptiveSampler) ShouldSample(method string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset window if needed
	if time.Since(a.lastWindowReset) > a.windowDuration {
		a.currentWindow = (a.currentWindow + 1) % len(a.requestCounts)
		a.requestCounts[a.currentWindow] = 0
		a.lastWindowReset = time.Now()
	}

	// Count request
	a.requestCounts[a.currentWindow]++

	// Calculate current rate
	currentRate := a.baseRate
	totalRequests := a.calculateTotalRequests()

	if totalRequests > a.requestWindow {
		// High volume - reduce sampling
		currentRate = a.minRate + (a.baseRate-a.minRate)*(float64(a.requestWindow)/float64(totalRequests))
	} else if totalRequests < a.requestWindow/2 {
		// Low volume - increase sampling
		currentRate = a.maxRate - (a.maxRate-a.baseRate)*(2*float64(totalRequests)/float64(a.requestWindow))
	}

	if currentRate < a.minRate {
		currentRate = a.minRate
	}
	if currentRate > a.maxRate {
		currentRate = a.maxRate
	}

	return a.rng.Float64() < currentRate
}

// calculateTotalRequests calculates total requests in the sliding window
func (a *AdaptiveSampler) calculateTotalRequests() int64 {
	total := int64(0)
	for _, count := range a.requestCounts {
		total += count
	}
	return total
}

// NoopRequestSampler represents a no-operation request sampler
type NoopRequestSampler struct{}

// NewNoopRequestSampler creates a new no-op request sampler
func NewNoopRequestSampler() RequestSampler {
	return &NoopRequestSampler{}
}

// ShouldSample always returns false
func (s *NoopRequestSampler) ShouldSample(method string) bool {
	return false
}

// GetSamplingRate returns 0
func (s *NoopRequestSampler) GetSamplingRate() float64 {
	return 0
}

// UpdateSamplingRate does nothing
func (s *NoopRequestSampler) UpdateSamplingRate(rate float64) {
	// No-op
}

// GetStats returns empty stats
func (s *NoopRequestSampler) GetStats() *SamplingStats {
	return &SamplingStats{
		TotalRequests:   0,
		SampledRequests: 0,
		SamplingRate:    0,
		LastUpdated:     time.Now(),
	}
}
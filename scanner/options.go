package scanner

// WithSingleBrewShotWeight sets a custom expected single shot weight
func WithSingleBrewShotWeight(weight float64) func(*Scanner) {
	return func(s *Scanner) {
		s.expectedSingleShotWeight = weight
	}
}

// WithDoubleBrewShotWeight sets a custom expected double shot weight
func WithDoubleBrewShotWeight(weight float64) func(*Scanner) {
	return func(s *Scanner) {
		s.expectedDoubleShotWeight = weight
	}
}

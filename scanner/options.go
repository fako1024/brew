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

// WithSingleShotBeansWeight sets a custom weight for the beans / grounds
// used for a single shot
func WithSingleShotBeansWeight(weight float64) func(*Scanner) {
	return func(s *Scanner) {
		s.singleShotBeansWeight = weight
	}
}

// WithDoubleShotBeansWeight sets a custom weight for the beans / grounds
// used for a double shot
func WithDoubleShotBeansWeight(weight float64) func(*Scanner) {
	return func(s *Scanner) {
		s.doubleShotBeansWeight = weight
	}
}

// WithGrindSetting sets a custom relative grinder setting:
// 0.0: finest
// 1.0: coarsest
func WithGrindSetting(setting float64) func(*Scanner) {
	return func(s *Scanner) {
		s.grindSetting = setting
	}
}

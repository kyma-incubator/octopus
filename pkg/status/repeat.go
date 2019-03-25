package status

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
)

type RepeatStrategy struct{}

func (s *RepeatStrategy) GetTestToRunConcurrently(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult {
	return s.getTest(suite, func(tr v1alpha1.TestResult) bool {
		return tr.DisableConcurrency == false
	})
}

func (s *RepeatStrategy) GetTestToRunSequentially(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult {
	return s.getTest(suite, func(tr v1alpha1.TestResult) bool {
		return tr.DisableConcurrency == true
	})
}

func (s *RepeatStrategy) getTest(suite v1alpha1.ClusterTestSuite, match func(tr v1alpha1.TestResult) bool) *v1alpha1.TestResult {
	repeat := s.getRepeat(suite)

	for _, tr := range suite.Status.Results {
		if !match(tr) {
			continue
		}
		if len(tr.Executions) < repeat {
			return &tr
		}
	}
	return nil
}

func (s *RepeatStrategy) IsApplicable(suite v1alpha1.ClusterTestSuite) bool {
	return suite.Spec.MaxRetries == 0
}

func (s *RepeatStrategy) getRepeat(suite v1alpha1.ClusterTestSuite) int {
	repeat := int(suite.Spec.Count)
	if repeat == 0 {
		repeat = 1 // TODO (aszecowka) until we don't set defaults in WebHook we normalize it here
	}
	return repeat
}

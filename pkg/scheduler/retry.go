package scheduler

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/api/core/v1"
)

type retryStrategy struct{}

func (r *retryStrategy) GetTestToRunConcurrently(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult {
	return r.getTest(suite, func(tr v1alpha1.TestResult) bool {
		return tr.DisabledConcurrency == false
	})
}

func (r *retryStrategy) GetTestToRunSequentially(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult {
	return r.getTest(suite, func(tr v1alpha1.TestResult) bool {
		return tr.DisabledConcurrency
	})
}

func (r *retryStrategy) getTest(suite v1alpha1.ClusterTestSuite, match func(tr v1alpha1.TestResult) bool) *v1alpha1.TestResult {
	for _, tr := range suite.Status.Results {
		if !match(tr) {
			continue
		}
		if len(tr.Executions) > int(suite.Spec.MaxRetries) {
			continue
		}

		hasSuccessful := false
		hasInProgress := false
		for _, ex := range tr.Executions {
			if ex.PodPhase == v1.PodSucceeded {
				hasSuccessful = true
				break
			} else if ex.PodPhase == v1.PodRunning || ex.PodPhase == v1.PodPending {
				hasInProgress = true
				break
			}
		}
		if hasSuccessful || hasInProgress {
			continue
		}
		return &tr
	}
	return nil
}

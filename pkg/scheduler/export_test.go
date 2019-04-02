package scheduler

import "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"

func (s *Service) GetNextToSchedule(suite v1alpha1.ClusterTestSuite) (*v1alpha1.TestResult, error) {
	return s.getNextToSchedule(suite)
}

package status

import (
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/consts"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type NowProvider func() time.Time

type Service struct {
	nowProvider NowProvider
}

func NewService(nowProvider NowProvider) *Service {
	return &Service{nowProvider: nowProvider}
}

func (s *Service) EnsureStatusIsUpToDate(suite v1alpha1.ClusterTestSuite, pods []v1.Pod) (*v1alpha1.TestSuiteStatus, error) {
	out := suite.Status.DeepCopy()
	for _, pod := range pods {
		for resultID, tr := range out.Results {
			if tr.Name == pod.Labels[consts.LabelTestDefName] && tr.Namespace == pod.Namespace {
				// find execution
				for execID, exec := range tr.Executions {
					if exec.ID == pod.Name {
						out.Results[resultID].Executions[execID].Status = s.getExecutionStatusFromPodPhase(pod)
						// TODO
					}
				}
			}
		}
	}
	// TODO what to do with incosistencies?
	return nil, nil
}

func (s *Service) getExecutionStatusFromPodPhase(pod v1.Pod) v1alpha1.TestExecutionStatus {
	switch pod.Status.Phase {
	case v1.PodPending:
		return v1alpha1.TestScheduled
	case v1.PodRunning:
		return v1alpha1.TestRunning
	case v1.PodSucceeded:
		return v1alpha1.TestSucceed
	case v1.PodFailed:
		return v1alpha1.TestFailed
	case v1.PodUnknown:
		return v1alpha1.TestNotYetScheduled

	}
	return v1alpha1.TestNotYetScheduled

}

func (s *Service) InitializeTests(suite v1alpha1.ClusterTestSuite, defs []v1alpha1.TestDefinition) (*v1alpha1.TestSuiteStatus, error) {
	out := suite.Status.DeepCopy()
	out.StartTime = &metav1.Time{Time: s.nowProvider()}
	if len(defs) == 0 {
		out.CompletionTime = &metav1.Time{Time: s.nowProvider()}
		s.SetSuiteCondition(out, v1alpha1.SuiteSucceed, "", "")
		return out, nil
	}
	s.SetSuiteCondition(out, v1alpha1.SuiteRunning, "", "")
	out.Results = make([]v1alpha1.TestResult, len(defs))
	for idx, def := range defs {
		out.Results[idx] = v1alpha1.TestResult{
			Name:       def.Name,
			Namespace:  def.Namespace,
			Executions: make([]v1alpha1.TestExecution, 0),
		}
	}

	return out, nil
}

func (s *Service) SetSuiteCondition(stat *v1alpha1.TestSuiteStatus, tp v1alpha1.TestSuiteConditionType, reason, msg string) {
	set := false
	for idx := 0; idx < len(stat.Conditions); idx++ {
		curr := &stat.Conditions[idx]
		if curr.Type == tp {
			curr.Status = v1alpha1.StatusTrue
			curr.Reason = reason
			curr.Message = msg
			set = true
		} else {
			curr.Status = v1alpha1.StatusFalse
			curr.Reason = ""
			curr.Message = ""
		}
	}
	if set {
		return
	}

	if stat.Conditions == nil {
		stat.Conditions = make([]v1alpha1.TestSuiteCondition, 0)
	}
	stat.Conditions = append(stat.Conditions, v1alpha1.TestSuiteCondition{
		Type:    tp,
		Status:  v1alpha1.StatusTrue,
		Reason:  reason,
		Message: msg,
	})
}

func (s *Service) IsUninitialized(suite v1alpha1.ClusterTestSuite) bool {
	if len(suite.Status.Conditions) == 0 {
		return true
	}

	for _, cond := range suite.Status.Conditions {
		if cond.Type == v1alpha1.SuiteUninitialized && cond.Status == v1alpha1.StatusTrue {
			return true
		}
	}
	return false
}

func (s *Service) IsFinished(suite v1alpha1.ClusterTestSuite) bool {
	return s.isConditionSet(suite, v1alpha1.SuiteError) ||
		s.isConditionSet(suite, v1alpha1.SuiteFailed) ||
		s.isConditionSet(suite, v1alpha1.SuiteSucceed)
}

func (s *Service) isConditionSet(suite v1alpha1.ClusterTestSuite, tp v1alpha1.TestSuiteConditionType) bool {
	for _, cond := range suite.Status.Conditions {
		if cond.Type == tp && cond.Status == v1alpha1.StatusTrue {
			return true
		}
	}
	return false
}

func (s *Service) GetNextToSchedule(suite v1alpha1.ClusterTestSuite) (*v1alpha1.TestResult, error) {
	// TODO no count, no retries
	for _, tr := range suite.Status.Results {
		if len(tr.Executions) == 0 {
			return &tr, nil
		}
	}
	return nil, nil

}

func (s *Service) MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName, testNs, podName string) (v1alpha1.TestSuiteStatus, error) {
	// TODO mark whole suite as started if needed
	for _, tr := range status.Results {
		if tr.Name == testName && tr.Namespace == testNs {
			tr.Executions = append(tr.Executions, v1alpha1.TestExecution{
				ID:        podName,
				StartTime: &metav1.Time{Time: s.nowProvider()},
				Status:    v1alpha1.TestScheduled,
			})

			return status, nil
		}
	}
	return v1alpha1.TestSuiteStatus{}, fmt.Errorf("cannot mark test as a scheduled [testName: %s, testNs: %s, podName: %s]", testName, testNs, podName)
}

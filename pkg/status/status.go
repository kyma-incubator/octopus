package status

import (
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		for idx, tr := range out.Results {
			if tr.Name == pod.Labels[v1alpha1.LabelKeyTestDefName] && tr.Namespace == pod.Namespace {
				// find execution
				for execID, exec := range tr.Executions {
					if exec.ID == pod.Name {
						prev := exec.PodPhase
						if pod.Status.Phase != prev {
							out.Results[idx].Executions[execID] = s.adjustTestExec(exec, pod)
						}
					}
				}
			}
		}
	}

	for idx, res := range out.Results {
		newState := s.calculateTestStatus(res, suite.Spec.MaxRetries, suite.Spec.Count)
		if res.Status != newState {
			out.Results[idx].Status = newState
		}
	}
	adjusted := s.adjustSuiteCondition(*out)
	out = &adjusted
	return out, nil
}

func (s *Service) adjustTestExec(exec v1alpha1.TestExecution, pod v1.Pod) v1alpha1.TestExecution {
	exec.PodPhase = pod.Status.Phase
	if exec.PodPhase == v1.PodSucceeded {
		exec.CompletionTime = &metav1.Time{Time: s.nowProvider()}
	} else if exec.PodPhase == v1.PodFailed {
		exec.CompletionTime = &metav1.Time{Time: s.nowProvider()}
		exec.Reason = pod.Status.Reason
		exec.Message = pod.Status.Message
	}
	return exec
}

func (s *Service) calculateTestStatus(tr v1alpha1.TestResult, maxRetries, count int64) v1alpha1.TestStatus {
	if len(tr.Executions) == 0 {
		return v1alpha1.TestNotYetScheduled
	}

	if maxRetries > 0 {
		var anySucceeded bool
		var anyRunning bool
		for _, exec := range tr.Executions {
			if exec.PodPhase == v1.PodSucceeded {
				anySucceeded = true
				break
			}
			if exec.PodPhase == v1.PodRunning || exec.PodPhase == v1.PodPending {
				anyRunning = true
				break
			}
		}
		if anySucceeded {
			return v1alpha1.TestSucceeded
		}
		if anyRunning {
			return v1alpha1.TestRunning
		}
		if len(tr.Executions) >= int(maxRetries+1) {
			return v1alpha1.TestFailed
		}
		return v1alpha1.TestRunning

	}

	if len(tr.Executions) < int(count) {
		return v1alpha1.TestRunning
	}

	var anyFailed bool
	for _, exec := range tr.Executions {
		switch exec.PodPhase {
		case v1.PodPending:
			return v1alpha1.TestRunning
		case v1.PodRunning:
			return v1alpha1.TestRunning
		case v1.PodFailed:
			anyFailed = true
		case v1.PodUnknown:
			return v1alpha1.TestRunning
		}
	}
	if anyFailed {
		return v1alpha1.TestFailed
	}
	return v1alpha1.TestSucceeded

}

func (s *Service) adjustSuiteCondition(stat v1alpha1.TestSuiteStatus) v1alpha1.TestSuiteStatus {
	prevCond := s.getSuiteCondition(stat)

	// TODO(aszecowka)(later) anySkipped, https://github.com/kyma-incubator/octopus/issues/10
	var anyNotScheduled, anyScheduled, anyRunning, anyUnknown, anyFailed bool
	var newCond v1alpha1.TestSuiteConditionType
	for _, res := range stat.Results {
		switch res.Status {
		case v1alpha1.TestNotYetScheduled:
			anyNotScheduled = true
		case v1alpha1.TestScheduled:
			anyScheduled = true

		case v1alpha1.TestRunning:
			anyRunning = true

		case v1alpha1.TestUnknown:
			anyUnknown = true

		case v1alpha1.TestFailed:
			anyFailed = true
		}
	}

	if anyRunning || anyNotScheduled || anyScheduled {
		newCond = v1alpha1.SuiteRunning
	} else if anyFailed {
		newCond = v1alpha1.SuiteFailed
	} else if anyUnknown {
		newCond = v1alpha1.SuiteError //TODO(aszecowka) later, should it be a error?
	} else {
		newCond = v1alpha1.SuiteSucceeded
	}

	if newCond == prevCond {
		return stat
	}
	stat.SetSuiteCondition(newCond, "", "")
	switch newCond {
	case v1alpha1.SuiteFailed:
		fallthrough
	case v1alpha1.SuiteSucceeded:
		fallthrough
	case v1alpha1.SuiteError:
		stat.CompletionTime = &metav1.Time{Time: s.nowProvider()}
	}

	return stat
}

func (s *Service) InitializeTests(suite v1alpha1.ClusterTestSuite, defs []v1alpha1.TestDefinition) (*v1alpha1.TestSuiteStatus, error) {
	out := suite.Status.DeepCopy()
	out.StartTime = &metav1.Time{Time: s.nowProvider()}
	if len(defs) == 0 {
		out.CompletionTime = &metav1.Time{Time: s.nowProvider()}
		out.SetSuiteCondition(v1alpha1.SuiteSucceeded, "", "")
		return out, nil
	}
	out.SetSuiteCondition(v1alpha1.SuiteRunning, "", "")
	out.Results = make([]v1alpha1.TestResult, len(defs))
	for idx, def := range defs {
		out.Results[idx] = v1alpha1.TestResult{
			Name:                def.Name,
			Namespace:           def.Namespace,
			Status:              v1alpha1.TestNotYetScheduled,
			Executions:          make([]v1alpha1.TestExecution, 0),
			DisabledConcurrency: def.Spec.DisableConcurrency,
		}
	}

	return out, nil
}

//
//func (s *Service) IsUninitialized(suite v1alpha1.ClusterTestSuite) bool {
//	if len(suite.Status.Conditions) == 0 {
//		return true
//	}
//
//	if s.isConditionSet(suite.Status, v1alpha1.SuiteUninitialized) {
//		return true
//	}
//
//	// if error occurred on initialization we treat suite as Uninitialized
//	for _, cond := range suite.Status.Conditions {
//		if cond.Type == v1alpha1.SuiteError && cond.Status == v1alpha1.StatusTrue && cond.Reason == v1alpha1.ReasonErrorOnInitialization {
//			return true
//		}
//	}
//	return false
//}

//func (s *Service) IsFinished(suite v1alpha1.ClusterTestSuite) bool {
//	return s.isConditionSet(suite.Status, v1alpha1.SuiteError) ||
//		s.isConditionSet(suite.Status, v1alpha1.SuiteFailed) ||
//		s.isConditionSet(suite.Status, v1alpha1.SuiteSucceeded)
//}

//func (s *Service) isConditionSet(stat v1alpha1.TestSuiteStatus, tp v1alpha1.TestSuiteConditionType) bool {
//	for _, cond := range stat.Conditions {
//		if cond.Type == tp && cond.Status == v1alpha1.StatusTrue {
//			return true
//		}
//	}
//	return false
//}

func (s *Service) getSuiteCondition(stat v1alpha1.TestSuiteStatus) v1alpha1.TestSuiteConditionType {
	for _, cond := range stat.Conditions {
		if cond.Status == v1alpha1.StatusTrue {
			return cond.Type
		}
	}
	return v1alpha1.SuiteUninitialized
}

//func (s *Service) GetExecutionsInProgress(suite v1alpha1.ClusterTestSuite) []v1alpha1.TestExecution {
//	out := make([]v1alpha1.TestExecution, 0)
//	for _, tr := range suite.Status.Results {
//		for _, ex := range tr.Executions {
//			if ex.PodPhase == v1.PodPending || ex.PodPhase == v1.PodRunning {
//				out = append(out, ex)
//			}
//		}
//	}
//	return out
//}

//func (s *Service) MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName, testNs, podName string) (v1alpha1.TestSuiteStatus, error) {
//	for idx, tr := range status.Results {
//		if tr.Name == testName && tr.Namespace == testNs {
//			status.Results[idx].Status = v1alpha1.TestScheduled
//			status.Results[idx].Executions = append(status.Results[idx].Executions, v1alpha1.TestExecution{
//				ID:        podName,
//				StartTime: &metav1.Time{Time: s.nowProvider()},
//			})
//
//			return status, nil
//		}
//	}
//	return v1alpha1.TestSuiteStatus{}, fmt.Errorf("cannot mark test as a scheduled [testName: %s, testNs: %s, podName: %s]", testName, testNs, podName)
//}

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
						prev := exec.PodPhase
						if pod.Status.Phase != prev {
							out.Results[resultID].Executions[execID] = s.adjustTestExec(exec, pod)
						}
					}
				}
			}
		}
	}

	for idx, res := range out.Results {
		newState := s.calculateTestStatus(res)
		if res.Status != newState {
			out.Results[idx].Status = newState
		}
	}
	adjusted := s.adjustSuiteCondition(*out)
	out = &adjusted
	// TODO what to do with incosistencies?
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

func (s *Service) calculateTestStatus(tr v1alpha1.TestResult) v1alpha1.TestStatus {
	if len(tr.Executions) == 0 {
		return v1alpha1.TestNotYetScheduled
	}

	var anyPending, anyRunning, anySucceeded, anyFailed, anyUnknown bool
	for _, exec := range tr.Executions {
		switch exec.PodPhase {
		case v1.PodPending:
			anyPending = true
		case v1.PodFailed:
			anyFailed = true
		case v1.PodRunning:
			anyRunning = true
		case v1.PodSucceeded:
			anySucceeded = true
		case v1.PodUnknown:
			anyUnknown = true
		}
	}
	if anyPending || anyRunning {
		return v1alpha1.TestRunning
	}
	if anyFailed {
		return v1alpha1.TestFailed
	}
	if anyUnknown {
		return v1alpha1.TestUnknown
	}
	if anySucceeded {
		return v1alpha1.TestSucceeded
	}

	return v1alpha1.TestUnknown

}

func (s *Service) adjustSuiteCondition(stat v1alpha1.TestSuiteStatus) v1alpha1.TestSuiteStatus {
	prevCond := s.getSuiteCondition(stat)

	// TODO anySkipped,
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
		newCond = v1alpha1.SuiteError //TODO
	} else {
		newCond = v1alpha1.SuiteSucceeded
	}

	if newCond == prevCond {
		return stat
	}
	s.SetSuiteCondition(&stat, newCond, "", "")
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
		s.SetSuiteCondition(out, v1alpha1.SuiteSucceeded, "", "")
		return out, nil
	}
	s.SetSuiteCondition(out, v1alpha1.SuiteRunning, "", "")
	out.Results = make([]v1alpha1.TestResult, len(defs))
	for idx, def := range defs {
		out.Results[idx] = v1alpha1.TestResult{
			Name:       def.Name,
			Namespace:  def.Namespace,
			Status:     v1alpha1.TestNotYetScheduled,
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
	return s.isConditionSet(suite.Status, v1alpha1.SuiteError) ||
		s.isConditionSet(suite.Status, v1alpha1.SuiteFailed) ||
		s.isConditionSet(suite.Status, v1alpha1.SuiteSucceeded)
}

func (s *Service) isConditionSet(stat v1alpha1.TestSuiteStatus, tp v1alpha1.TestSuiteConditionType) bool {
	for _, cond := range stat.Conditions {
		if cond.Type == tp && cond.Status == v1alpha1.StatusTrue {
			return true
		}
	}
	return false
}

func (s *Service) getSuiteCondition(stat v1alpha1.TestSuiteStatus) v1alpha1.TestSuiteConditionType {
	for _, cond := range stat.Conditions {
		if cond.Status == v1alpha1.StatusTrue {
			return cond.Type
		}
	}
	return v1alpha1.SuiteUninitialized
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
	for idx, tr := range status.Results {
		if tr.Name == testName && tr.Namespace == testNs {
			status.Results[idx].Status = v1alpha1.TestScheduled
			status.Results[idx].Executions = append(status.Results[idx].Executions, v1alpha1.TestExecution{
				ID:        podName,
				StartTime: &metav1.Time{Time: s.nowProvider()},
			})

			return status, nil
		}
	}
	return v1alpha1.TestSuiteStatus{}, fmt.Errorf("cannot mark test as a scheduled [testName: %s, testNs: %s, podName: %s]", testName, testNs, podName)
}

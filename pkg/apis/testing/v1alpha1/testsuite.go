package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"fmt"
	"k8s.io/api/core/v1"
)

func (in *TestSuiteStatus) SetSuiteCondition(tp TestSuiteConditionType, reason, msg string) {
	set := false
	for idx := 0; idx < len(in.Conditions); idx++ {
		curr := &in.Conditions[idx]
		if curr.Type == tp {
			curr.Status = StatusTrue
			curr.Reason = reason
			curr.Message = msg
			set = true
		} else {
			curr.Status = StatusFalse
			curr.Reason = ""
			curr.Message = ""
		}
	}
	if set {
		return
	}

	if in.Conditions == nil {
		in.Conditions = make([]TestSuiteCondition, 0)
	}
	in.Conditions = append(in.Conditions, TestSuiteCondition{
		Type:    tp,
		Status:  StatusTrue,
		Reason:  reason,
		Message: msg,
	})
}

func (status *TestSuiteStatus) MarkAsScheduled(testName, testNs, podName string, now time.Time) (*TestSuiteStatus, error) {
	for idx, tr := range status.Results {
		if tr.Name == testName && tr.Namespace == testNs {
			status.Results[idx].Status = TestScheduled
			status.Results[idx].Executions = append(status.Results[idx].Executions, TestExecution{
				ID:        podName,
				StartTime: &metav1.Time{Time: now},
			})

			return status, nil
		}
	}
	return &TestSuiteStatus{}, fmt.Errorf("cannot mark test as a scheduled [testName: %s, testNs: %s, podName: %s]", testName, testNs, podName)
}

func (suite *ClusterTestSuite) GetExecutionsInProgress() []TestExecution {
	out := make([]TestExecution, 0)
	for _, tr := range suite.Status.Results {
		for _, ex := range tr.Executions {
			if ex.PodPhase == v1.PodPending || ex.PodPhase == v1.PodRunning {
				out = append(out, ex)
			}
		}
	}
	return out
}

func (suite *ClusterTestSuite) IsFinished() bool {
	return suite.Status.isConditionSet(SuiteError) ||
		suite.Status.isConditionSet(SuiteFailed) ||
		suite.Status.isConditionSet(SuiteSucceeded)
}

func (stat *TestSuiteStatus) isConditionSet(tp TestSuiteConditionType) bool {
	for _, cond := range stat.Conditions {
		if cond.Type == tp && cond.Status == StatusTrue {
			return true
		}
	}
	return false
}


func (suite *ClusterTestSuite) IsUninitialized() bool {
	if len(suite.Status.Conditions) == 0 {
		return true
	}

	if suite.Status.isConditionSet(SuiteUninitialized) {
		return true
	}

	// if error occurred on initialization we treat suite as Uninitialized
	for _, cond := range suite.Status.Conditions {
		if cond.Type == SuiteError && cond.Status == StatusTrue && cond.Reason == ReasonErrorOnInitialization {
			return true
		}
	}
	return false
}
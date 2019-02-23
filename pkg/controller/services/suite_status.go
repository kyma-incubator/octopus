package services

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type NowProvider interface {
	Now() time.Time
}

type Status struct {
	nowProvider NowProvider
}

func (s *Status) IsUninitialized(suite *v1alpha1.TestSuite) bool {
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

func (s *Status) isConditionSet(suite *v1alpha1.TestSuite, tp v1alpha1.TestSuiteConditionType) bool {
	for _, cond := range suite.Status.Conditions {
		if cond.Type == tp && cond.Status == v1alpha1.StatusTrue {
			return true
		}
	}
	return false
}

func (s *Status) setCondition(suite *v1alpha1.TestSuite, tp v1alpha1.TestSuiteConditionType, val v1alpha1.Status) {
	for _, cond := range suite.Status.Conditions {
		if cond.Type == tp {
			cond.Status = val
			return
		}
	}
	if suite.Status.Conditions == nil {
		// TODO not sure if this condition is possible
		suite.Status.Conditions = make([]v1alpha1.TestSuiteCondition, 0)
	}

	suite.Status.Conditions = append(suite.Status.Conditions, v1alpha1.TestSuiteCondition{
		Type:   tp,
		Status: val,
	})
}

func (s *Status) IsFinished(suite *v1alpha1.TestSuite) bool {
	return s.isConditionSet(suite, v1alpha1.SuiteError) ||
		s.isConditionSet(suite, v1alpha1.SuiteFailed) ||
		s.isConditionSet(suite, v1alpha1.SuiteSucceed)
}

func (s *Status) InitializeTests(suite *v1alpha1.TestSuite, testsDef []v1alpha1.TestDefinition) error {
	suite.Status.StartTime = &v1.Time{Time: s.nowProvider.Now()}
	s.setCondition(suite, v1alpha1.SuiteUninitialized, v1alpha1.StatusFalse)

	if len(testsDef) == 0 {
		suite.Status.CompletionTime = &v1.Time{Time: s.nowProvider.Now()}
		s.setCondition(suite, v1alpha1.SuiteSucceed, v1alpha1.StatusTrue)
		return nil
	}

	s.setCondition(suite, v1alpha1.SuiteRunning, v1alpha1.StatusTrue)
	suite.Status.Results = make([]v1alpha1.TestResult, 0)
	for _, td := range testsDef {
		tr := v1alpha1.TestResult{
			Name:      td.Name,
			Namespace: td.Namespace,
		}
		s.SetTestResultCondition(&tr, v1alpha1.TestPending, v1alpha1.StatusTrue)
		suite.Status.Results = append(suite.Status.Results, tr)
	}

	return nil
}


func (s *Status) SetTestResultCondition(tr *v1alpha1.TestResult, tp v1alpha1.TestResultConditionType, value v1alpha1.Status) {
	for _, cond := range tr.Conditions {
		if cond.Type == tp {
			cond.Status = value
		}
	}

	if tr.Conditions == nil {
		tr.Conditions = make([]v1alpha1.TestResultCondition, 0)
	}

	tr.Conditions = append(tr.Conditions, v1alpha1.TestResultCondition{
		Type:   tp,
		Status: value,
	})

}

func (s *Status) GetTestResultByCondition(suite *v1alpha1.TestSuite, tp v1alpha1.TestResultConditionType) ([]v1alpha1.TestResult, error) {
	var out []v1alpha1.TestResult
	for _, testResult := range suite.Status.Results {
		for _, cond := range testResult.Conditions {
			if cond.Type == tp && cond.Status == v1alpha1.StatusTrue {
				out = append(out, testResult)
				break
			}
		}
	}
	return out, nil
}

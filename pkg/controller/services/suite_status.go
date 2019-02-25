package services

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type NowProvider interface {
	Now() time.Time
}

func NewStatus(np NowProvider) *Status {
	return &Status{
		nowProvider: np,
	}
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

func (s *Status) setCondition(suite *v1alpha1.TestSuite, tp v1alpha1.TestSuiteConditionType) {
	found := false
	for idx, _ := range suite.Status.Conditions {
		if suite.Status.Conditions[idx].Type == tp {
			suite.Status.Conditions[idx].Status = v1alpha1.StatusTrue
			found = true
		} else {
			suite.Status.Conditions[idx].Status = v1alpha1.StatusFalse
		}
	}
	if found {
		return
	}
	if suite.Status.Conditions == nil {
		// TODO not sure if this condition is possible
		suite.Status.Conditions = make([]v1alpha1.TestSuiteCondition, 0)
	}

	suite.Status.Conditions = append(suite.Status.Conditions, v1alpha1.TestSuiteCondition{
		Type:   tp,
		Status: v1alpha1.StatusTrue,
	})
}

func (s *Status) IsFinished(suite *v1alpha1.TestSuite) bool {
	return s.isConditionSet(suite, v1alpha1.SuiteError) ||
		s.isConditionSet(suite, v1alpha1.SuiteFailed) ||
		s.isConditionSet(suite, v1alpha1.SuiteSucceed)
}

func (s *Status) InitializeTests(suite *v1alpha1.TestSuite, testsDef []v1alpha1.TestDefinition) error {
	suite.Status.StartTime = &v1.Time{Time: s.nowProvider.Now()}

	if len(testsDef) == 0 {
		suite.Status.CompletionTime = &v1.Time{Time: s.nowProvider.Now()}
		s.setCondition(suite, v1alpha1.SuiteSucceed)
		return nil
	}

	s.setCondition(suite, v1alpha1.SuiteRunning)
	suite.Status.Results = make([]v1alpha1.TestResult, 0)
	for _, td := range testsDef {
		tr := v1alpha1.TestResult{
			Name:      td.Name,
			Namespace: td.Namespace,
		}
		s.SetTestResultCondition(&tr, v1alpha1.TestPending)
		suite.Status.Results = append(suite.Status.Results, tr)
	}

	return nil
}

func (s *Status) SetTestResultCondition(tr *v1alpha1.TestResult, tp v1alpha1.TestResultConditionType) {
	found := false
	for idx, _ := range tr.Conditions {
		if tr.Conditions[idx].Type == tp {
			tr.Conditions[idx].Status = v1alpha1.StatusTrue
			found = true
		} else {
			tr.Conditions[idx].Status = v1alpha1.StatusFalse
		}
	}
	if found {
		return
	}

	if tr.Conditions == nil {
		tr.Conditions = make([]v1alpha1.TestResultCondition, 0)
	}

	tr.Conditions = append(tr.Conditions, v1alpha1.TestResultCondition{
		Type:   tp,
		Status: v1alpha1.StatusTrue,
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

func (s *Status) EnsureStatusIsUpToDate(suite *v1alpha1.TestSuite, pods []v12.Pod) error {
	for _, pod := range pods {
		id := pod.Name
		for idx, tr := range suite.Status.Results {
			if tr.ID == id {
				switch pod.Status.Phase {
				// TODO s.SetTestResultCondition(&tr,v1alpha1.TestError, v1alpha1.StatusTrue)
				case v12.PodPending:
					s.SetTestResultCondition(&suite.Status.Results[idx], v1alpha1.TestRunning)
				case v12.PodRunning:
					s.SetTestResultCondition(&suite.Status.Results[idx], v1alpha1.TestRunning)
				case v12.PodSucceeded:
					s.SetTestResultCondition(&suite.Status.Results[idx], v1alpha1.TestSucceed)
				case v12.PodFailed:
					s.SetTestResultCondition(&suite.Status.Results[idx], v1alpha1.TestFailed)
				case v12.PodUnknown:
					//TODO
				default:
					return errors.New("unsupported pod phase")
				}
				break
			}
		}
	}

	running, err := s.GetTestResultByCondition(suite, v1alpha1.TestRunning)
	if err != nil {
		return err
	}
	failed, err := s.GetTestResultByCondition(suite, v1alpha1.TestFailed)
	if err != nil {
		return err
	}

	errored, err := s.GetTestResultByCondition(suite, v1alpha1.TestError)
	if err != nil {
		return err
	}

	pending, err := s.GetTestResultByCondition(suite, v1alpha1.TestPending)
	if err != nil {
		return err
	}

	if len(running)+len(pending) > 0 {
		s.setCondition(suite, v1alpha1.SuiteRunning)
		return nil
	}
	if len(errored) > 0 {
		s.setCondition(suite, v1alpha1.SuiteError)
		return nil
	}

	if len(failed) > 0 {
		s.setCondition(suite, v1alpha1.SuiteFailed)
		return nil
	}

	s.setCondition(suite, v1alpha1.SuiteSucceed)
	return nil
}

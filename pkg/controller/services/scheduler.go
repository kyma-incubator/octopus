package services

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type StatusProvider interface {
	GetTestResultByCondition(suite *v1alpha1.TestSuite, tp v1alpha1.TestResultConditionType) ([]v1alpha1.TestResult, error)
	SetTestResultCondition(tr *v1alpha1.TestResult, tp v1alpha1.TestResultConditionType, value v1alpha1.Status)
}

type Scheduler struct {
	reader         client.Reader
	writer         client.Writer
	statusProvider StatusProvider
	scheme         *runtime.Scheme
}

func (s *Scheduler) tryScheduleTest(suite *v1alpha1.TestSuite) (*v1.ObjectReference, error) {
	ctx := context.TODO()

	if !s.canScheduleNext(suite) {
		return nil, nil
	}
	toSched, err := s.getNextToSchedule(suite)
	if err != nil {
		return nil, err
	}
	if toSched == nil {
		return nil, nil
	}
	var def v1alpha1.TestDefinition
	if err := s.reader.Get(ctx, client.ObjectKey{Namespace: toSched.Name, Name: toSched.Namespace}, &def); err != nil {
		// TODO avoid infinte loop for cases when always error is returned
		return nil, err
	}
	pod, err := s.startPod(suite, def)
	if err != nil {
		return nil, err
	}
	s.statusProvider.SetTestResultCondition(toSched, v1alpha1.TestPending, v1alpha1.StatusFalse)
	s.statusProvider.SetTestResultCondition(toSched, v1alpha1.TestRunning, v1alpha1.StatusTrue)
	return reference.GetReference(s.scheme, pod)

}

func (s *Scheduler) canScheduleNext(suite *v1alpha1.TestSuite) bool {
	return true // TODO huge simplification
}

func (s *Scheduler) getNextToSchedule(suite *v1alpha1.TestSuite) (*v1alpha1.TestResult, error) {
	tests, err := s.statusProvider.GetTestResultByCondition(suite, v1alpha1.TestPending)
	if err != nil {
		return nil, err
	}
	if len(tests) == 0 {
		return nil, nil
	}
	return &tests[0], nil
}

func (s *Scheduler) startPod(suite *v1alpha1.TestSuite, def v1alpha1.TestDefinition) (*v1.Pod, error) {
	p := &v1.Pod{}
	p.Spec = def.Spec.PodSpec
	p.GenerateName = "octopus-testing-pod-"
	p.Namespace = def.Namespace
	p.Labels = make(map[string]string)
	p.Spec.RestartPolicy = v1.RestartPolicyNever
	if err := controllerutil.SetControllerReference(suite, p, s.scheme); err != nil {
		return nil, err
	}

	return p, s.writer.Create(context.TODO(), p)
}

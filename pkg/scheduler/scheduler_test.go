package scheduler_test

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/consts"
	"github.com/kyma-incubator/octopus/pkg/scheduler"
	"github.com/kyma-incubator/octopus/pkg/scheduler/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestTryScheduleHappyPath(t *testing.T) {
	// GIVEN
	givenTr := givenTestResult()
	uninitializedSuite := givenUninitializedSuite(givenTr)
	givenTd := givenTestDefinition()

	scheduledSuite := uninitializedSuite.DeepCopy()
	scheduledSuite.Status.Conditions[0].Type = v1alpha1.SuiteRunning
	scheduledSuite.Status.Results[0].Executions = []v1alpha1.TestExecution{{ID: "octopus-testing-pod-0", Status: v1alpha1.TestRunning}}

	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)
	mockStatusProvider.On("GetNextToSchedule", uninitializedSuite).Return(&givenTr, nil).Once()
	mockStatusProvider.On("MarkAsScheduled", uninitializedSuite.Status, "test-name", "test-namespace", mock.Anything).Return(scheduledSuite.Status, nil)

	fakeCli, err := getFakeClient(&givenTd)
	require.NoError(t, err)
	wrappedWriter := podWriterExtended{cli: fakeCli}

	sut := scheduler.NewService(mockStatusProvider, fakeCli, &wrappedWriter)

	// WHEN
	pod, status, err := sut.TrySchedule(uninitializedSuite)
	// THEN
	require.NoError(t, err)
	assert.NotNil(t, pod)
	assert.Equal(t, scheduledSuite.Status, *status)

	var actualPodList v12.PodList
	require.NoError(t, fakeCli.List(context.TODO(), &client.ListOptions{Namespace: "test-namespace"}, &actualPodList))
	assert.Contains(t, actualPodList.Items, *pod)

	assert.Equal(t, "octopus-testing-pod-0", pod.Name)
	assert.Equal(t, "test-namespace", pod.Namespace)
	assert.Len(t, pod.Spec.Containers, 1)
	assert.Equal(t, "alpine", pod.Spec.Containers[0].Image)
	assert.Equal(t, v12.RestartPolicyNever, pod.Spec.RestartPolicy)
	labelCreatedByOctopus := pod.Labels[consts.LabelKeyCreatedByOctopus]
	assert.Equal(t, "true", labelCreatedByOctopus)
	labelSuiteName := pod.Labels[consts.LabelKeySuiteName]
	assert.Equal(t, "test-all", labelSuiteName)
	labelsTestName := pod.Labels[consts.LabelTestDefName]
	assert.Equal(t, "test-name", labelsTestName)

}

func getFakeClient(initOjbects ...runtime.Object) (client.Client, error) {
	sch := scheme.Scheme
	if err := v1alpha1.SchemeBuilder.AddToScheme(sch); err != nil {
		return nil, err
	}

	fakeCli := fake.NewFakeClientWithScheme(sch, initOjbects...)
	return fakeCli, nil
}

func givenTestResult() v1alpha1.TestResult {
	givenTr := v1alpha1.TestResult{
		Name:      "test-name",
		Namespace: "test-namespace",
	}
	return givenTr
}

func givenTestDefinition() v1alpha1.TestDefinition {
	givenTd := v1alpha1.TestDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.TestDefinitionSpec{
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Image: "alpine",
						},
					},
				},
			},
		},
	}
	return givenTd
}

func givenUninitializedSuite(givenTr v1alpha1.TestResult) v1alpha1.ClusterTestSuite {
	uninitializedSuite := v1alpha1.ClusterTestSuite{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-all",
		},
		Status: v1alpha1.TestSuiteStatus{
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Status: v1alpha1.StatusTrue,
					Type:   v1alpha1.SuiteUninitialized,
				},
			},
			Results: []v1alpha1.TestResult{
				givenTr,
			},
		},
	}
	return uninitializedSuite
}

func TestTryScheduleNoMoreTests(t *testing.T) {
	// GIVEN
	fakeCli, err := getFakeClient()
	require.NoError(t, err)
	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)
	mockStatusProvider.On("GetNextToSchedule", mock.Anything).Return(nil, nil)
	sut := scheduler.NewService(mockStatusProvider, fakeCli, fakeCli)
	// WHEN
	actualPod, actualStatus, err := sut.TrySchedule(givenUninitializedSuite(givenTestResult()))
	// THEN
	assert.NoError(t, err)
	assert.Nil(t, actualPod)
	assert.Nil(t, actualStatus)

}

func TestTryScheduleErrorOnGettingNextTest(t *testing.T) {
	//TODO
}

func TestTryScheduleErrorOnGettingTestDef(t *testing.T) {
	//TODO
}

func TestTryScheduleErrorOnSchedulingPod(t *testing.T) {
	//TODO
}

func TestTryScheduleErrorOnUpdatingStatus(t *testing.T) {
	//TODO
}

type podWriterExtended struct {
	cli client.Writer
	idx int
}

func (w *podWriterExtended) Create(ctx context.Context, obj runtime.Object) error {
	om, ok := obj.(*v12.Pod)
	if !ok {
		return errors.New("unsupported type")
	}
	if om.GenerateName != "" {
		om.Name = fmt.Sprintf("%s%d", om.GenerateName, w.idx)
		w.idx++
	}
	return w.cli.Create(ctx, obj)
}

func (w *podWriterExtended) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	return w.cli.Delete(ctx, obj, opts...)
}

func (w *podWriterExtended) Update(ctx context.Context, obj runtime.Object) error {
	return w.cli.Update(ctx, obj)
}

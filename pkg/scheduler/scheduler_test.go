package scheduler_test

import (
	"context"
	"fmt"
	"testing"

	rlog "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
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
)

func TestTryScheduleHappyPath(t *testing.T) {
	// GIVEN
	givenTr := givenTestResult()
	uninitializedSuite := givenUninitializedSuite(givenTr)
	givenTd := givenTestDefinition()

	scheduledSuite := uninitializedSuite.DeepCopy()
	scheduledSuite.Status.Conditions[0].Type = v1alpha1.SuiteRunning
	scheduledSuite.Status.Results[0].Executions = []v1alpha1.TestExecution{{ID: "oct-tp-test-all-test-name-0", PodPhase: v12.PodRunning}}

	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)
	mockStatusProvider.On("GetExecutionsInProgress", uninitializedSuite).Return(nil).Once()
	mockStatusProvider.On("MarkAsScheduled", uninitializedSuite.Status, "test-name", "test-namespace", mock.Anything).Return(scheduledSuite.Status, nil)

	fakeCli, sch, err := getFakeClient(&givenTd)

	require.NoError(t, err)

	sut := scheduler.NewService(mockStatusProvider, fakeCli, fakeCli, sch, rlog.NullLogger{})

	// WHEN
	pod, status, err := sut.TrySchedule(uninitializedSuite)
	// THEN
	require.NoError(t, err)
	assert.NotNil(t, pod)
	assert.Equal(t, scheduledSuite.Status, *status)

	var actualPodList v12.PodList
	require.NoError(t, fakeCli.List(context.TODO(), &client.ListOptions{Namespace: "test-namespace"}, &actualPodList))
	assert.Contains(t, actualPodList.Items, *pod)

	assert.Equal(t, "oct-tp-test-all-test-name-0", pod.Name)
	assert.Equal(t, "test-namespace", pod.Namespace)
	assert.Len(t, pod.Spec.Containers, 1)
	assert.Equal(t, "alpine", pod.Spec.Containers[0].Image)
	assert.Equal(t, v12.RestartPolicyNever, pod.Spec.RestartPolicy)
	labelCreatedByOctopus := pod.Labels[v1alpha1.LabelKeyCreatedByOctopus]
	assert.Equal(t, "true", labelCreatedByOctopus)
	labelSuiteName := pod.Labels[v1alpha1.LabelKeySuiteName]
	assert.Equal(t, "test-all", labelSuiteName)
	labelsTestName := pod.Labels[v1alpha1.LabelKeyTestDefName]
	assert.Equal(t, "test-name", labelsTestName)
	require.Len(t, pod.OwnerReferences, 1)
	require.NotNil(t, pod.OwnerReferences[0].Controller)
	assert.True(t, *pod.OwnerReferences[0].Controller)
	assert.Equal(t, "test-all", pod.OwnerReferences[0].Name)
}

func TestTryScheduleErrorOnGettingNextTest(t *testing.T) {
	// GIVEN
	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)

	mockLogger := &automock.Logger{}
	defer mockLogger.AssertExpectations(t)

	suite := v1alpha1.ClusterTestSuite{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-all",
		},
		Spec: v1alpha1.TestSuiteSpec{
			Concurrency: 1,
			Count:       1,
			//currently MaxRetries is not supported, so it should return error
			MaxRetries: 100,
		},
	}

	mockStatusProvider.On("GetExecutionsInProgress", suite).Return(nil)
	mockLogger.ExpectLoggedWithValues("suite", "test-all")
	mockLogger.ExpectLoggedOnError(errors.New("cannot find test selector strategy that is applicable for suite [test-all]"), "No applicable strategy")

	sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
	// WHEN
	_, _, err := sut.TrySchedule(suite)
	// THEN
	require.EqualError(t, err, "while getting next to schedule: cannot find test selector strategy that is applicable for suite [test-all]")
}

func TestTryScheduleNoTestToExecuteNow(t *testing.T) {
	// GIVEN
	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)

	mockLogger := &automock.Logger{}
	defer mockLogger.AssertExpectations(t)
	mockLogger.ExpectLoggedWithValues("suite", "test-all")
	mockLogger.ExpectLoggedOnInfo("No tests to execute right now")

	suite := v1alpha1.ClusterTestSuite{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-all",
		},
		Spec: v1alpha1.TestSuiteSpec{
			Concurrency: 1,
			Count:       1,
		},
	}
	mockStatusProvider.On("GetExecutionsInProgress", suite).Return(nil)
	sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
	// WHEN
	actualPod, actualStatus, err := sut.TrySchedule(suite)
	// THEN
	require.NoError(t, err)
	assert.Nil(t, actualPod)
	assert.Nil(t, actualStatus)
}

func TestTryScheduleErrorOnGettingTestDef(t *testing.T) {
	// GIVEN
	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)

	mockStatusProvider.On("GetExecutionsInProgress", mock.Anything).Return(nil)

	mockLogger := &automock.Logger{}
	defer mockLogger.AssertExpectations(t)
	mockLogger.ExpectLoggedWithValues("suite", "test-all")
	givenTr := givenTestResult()
	uninitializedSuite := givenUninitializedSuite(givenTr)

	fakeCli, sch, err := getFakeClient()
	require.NoError(t, err)

	sut := scheduler.NewService(mockStatusProvider, fakeCli, nil, sch, mockLogger)
	// WHEN
	_, _, err = sut.TrySchedule(uninitializedSuite)
	// THEN
	require.EqualError(t, err, "while getting test definition [name: test-name, namespace: test-namespace]: testdefinitions.testing.kyma-project.io \"test-name\" not found")
}

func TestTryScheduleErrorOnSchedulingPod(t *testing.T) {
	// GIVEN
	givenTr := givenTestResult()
	uninitializedSuite := givenUninitializedSuite(givenTr)
	givenTd := givenTestDefinition()

	scheduledSuite := uninitializedSuite.DeepCopy()
	scheduledSuite.Status.Conditions[0].Type = v1alpha1.SuiteRunning
	scheduledSuite.Status.Results[0].Executions = []v1alpha1.TestExecution{{ID: "octopus-testing-pod-0", PodPhase: v12.PodRunning}}

	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)
	mockStatusProvider.On("GetExecutionsInProgress", uninitializedSuite).Return(nil).Once()
	// TODO (aszecowka): later we should mark somehow on test suite, that error occurred

	fakeCli, sch, err := getFakeClient(&givenTd)
	require.NoError(t, err)
	mockWriter := &automock.Client{}
	defer mockWriter.AssertExpectations(t)
	mockWriter.On("Create", mock.Anything, mock.Anything).Return(errors.New("some error"))

	sut := scheduler.NewService(mockStatusProvider, fakeCli, mockWriter, sch, rlog.NullLogger{})

	// WHEN
	_, _, err = sut.TrySchedule(uninitializedSuite)
	// THEN
	assert.EqualError(t, err, "while creating testing pod for suite [test-all] and test definition [name: test-name, namespace: test-namespace]: some error")
}

func TestTryScheduleErrorOnUpdatingStatus(t *testing.T) {
	// GIVEN
	givenTr := givenTestResult()
	uninitializedSuite := givenUninitializedSuite(givenTr)
	givenTd := givenTestDefinition()

	scheduledSuite := uninitializedSuite.DeepCopy()
	scheduledSuite.Status.Conditions[0].Type = v1alpha1.SuiteRunning
	scheduledSuite.Status.Results[0].Executions = []v1alpha1.TestExecution{{ID: "oct-tp-test-all-test-name-0", PodPhase: v12.PodRunning}}

	mockStatusProvider := &automock.StatusProvider{}
	defer mockStatusProvider.AssertExpectations(t)
	mockStatusProvider.On("GetExecutionsInProgress", uninitializedSuite).Return(nil).Once()
	mockStatusProvider.On("MarkAsScheduled", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(v1alpha1.TestSuiteStatus{}, errors.New("some error"))

	fakeCli, sch, err := getFakeClient(&givenTd)
	require.NoError(t, err)

	sut := scheduler.NewService(mockStatusProvider, fakeCli, fakeCli, sch, rlog.NullLogger{})

	// WHEN
	_, _, err = sut.TrySchedule(uninitializedSuite)
	// THEN
	require.EqualError(t, err, "while marking suite [test-all] as Scheduled: some error")
}

func TestGetNextToSchedule(t *testing.T) {
	t.Run("returns nil if number of running tests is equal to concurrency level", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "all-tests",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Concurrency: 3,
				Count:       1,
			},
		}
		mockStatusProvider := &automock.StatusProvider{}
		mockStatusProvider.On("GetExecutionsInProgress", suite).Return([]v1alpha1.TestExecution{
			{ID: "id-111"}, {ID: "id-222"}, {ID: "id-333"},
		})

		mockLogger := &automock.Logger{}
		mockLogger.ExpectLoggedWithValues("suite", "all-tests")
		mockLogger.ExpectLoggedOnInfo("Cannot get next test to schedule, max concurrency reached", "running", 3, "concurrency", int64(3))
		defer mockLogger.AssertExpectations(t)
		defer mockStatusProvider.AssertExpectations(t)
		sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)

		// WHEN
		actual, err := sut.GetNextToSchedule(suite)
		// THEN
		require.NoError(t, err)
		require.Nil(t, actual)
	})

	t.Run("returns error if strategy for finding next to schedule is not yet implemented", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "all-tests",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Concurrency: 1,
				MaxRetries:  10,
			},
		}
		mockLogger := &automock.Logger{}
		defer mockLogger.AssertExpectations(t)
		mockLogger.ExpectLoggedWithValues("suite", "all-tests")
		mockLogger.ExpectLoggedOnError(fmt.Errorf("cannot find test selector strategy that is applicable for suite [%s]", "all-tests"), "No applicable strategy")
		mockStatusProvider := &automock.StatusProvider{}
		mockStatusProvider.On("GetExecutionsInProgress", mock.Anything).Return([]v1alpha1.TestExecution{}).Once()
		defer mockStatusProvider.AssertExpectations(t)
		sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
		// WHEN
		_, err := sut.GetNextToSchedule(suite)
		// THEN
		require.Error(t, err)

	})

	t.Run("returns concurrent test before sequential tests", func(t *testing.T) {
		// GIVEN
		mockStatusProvider := &automock.StatusProvider{}
		defer mockStatusProvider.AssertExpectations(t)
		mockStatusProvider.On("GetExecutionsInProgress", mock.Anything).Return([]v1alpha1.TestExecution{{ID: "id-222"}})
		mockLogger := &automock.Logger{}
		mockLogger.ExpectLoggedWithValues("suite", "test-all")
		defer mockLogger.AssertExpectations(t)
		sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
		// WHEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Concurrency: 2,
				Count:       1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test-1",
						DisabledConcurrency: true,
					},
					{
						Name:                "test-2",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{
								ID: "id-222",
							},
						},
					},
					{
						Name:                "test-3",
						DisabledConcurrency: false,
					},
				},
			},
		}
		actual, err := sut.GetNextToSchedule(suite)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, "test-3", actual.Name)
	})

	t.Run("returns sequential test", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Concurrency: 2,
				Count:       1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test-1",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-111",
								PodPhase: v12.PodSucceeded,
							},
						},
					},
					{
						Name:                "test-2",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-222",
								PodPhase: v12.PodSucceeded,
							},
						},
					},
					{
						Name:                "test-3",
						DisabledConcurrency: true,
						Executions:          []v1alpha1.TestExecution{},
					},
				},
			},
		}
		mockStatusProvider := &automock.StatusProvider{}
		defer mockStatusProvider.AssertExpectations(t)

		mockLogger := &automock.Logger{}
		defer mockLogger.AssertExpectations(t)

		mockStatusProvider.On("GetExecutionsInProgress", suite).Return(nil)
		mockLogger.ExpectLoggedWithValues("suite", "test-all")

		sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
		// WHEN
		actual, err := sut.GetNextToSchedule(suite)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, "test-3", actual.Name)
	})

	t.Run("returns nil if there is pending sequential test but other test is running", func(t *testing.T) {
		// GIVEN
		mockStatusProvider := &automock.StatusProvider{}
		defer mockStatusProvider.AssertExpectations(t)

		mockLogger := &automock.Logger{}
		defer mockLogger.AssertExpectations(t)

		mockStatusProvider.On("GetExecutionsInProgress", mock.Anything).Return([]v1alpha1.TestExecution{
			{ID: "id-222"},
		})

		mockLogger.ExpectLoggedWithValues("suite", "test-all")
		mockLogger.ExpectLoggedOnInfo("No tests to execute right now")
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Concurrency: 2,
				Count:       1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test-1",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-111",
								PodPhase: v12.PodSucceeded,
							},
						},
					},
					{
						Name:                "test-2",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-222",
								PodPhase: v12.PodRunning,
							},
						},
					},
					{
						Name:                "test-3",
						DisabledConcurrency: true,
					},
				},
			},
		}

		sut := scheduler.NewService(mockStatusProvider, nil, nil, nil, mockLogger)
		// WHEN
		actual, err := sut.GetNextToSchedule(suite)
		// THEN
		require.NoError(t, err)
		require.Nil(t, actual)
	})

}

// fake clients which supports Occtopus CRDs
func getFakeClient(initObjects ...runtime.Object) (client.Client, *runtime.Scheme, error) {
	sch := scheme.Scheme
	if err := v1alpha1.SchemeBuilder.AddToScheme(sch); err != nil {
		return nil, nil, err
	}

	fakeCli := fake.NewFakeClientWithScheme(sch, initObjects...)
	return fakeCli, sch, nil
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
		Spec: v1alpha1.TestSuiteSpec{
			Count:       1,
			Concurrency: 1,
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

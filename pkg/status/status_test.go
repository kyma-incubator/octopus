package status_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsUninitialized(t *testing.T) {
	sut := status.Service{}

	t.Run("return true when empty", func(t *testing.T) {
		assert.True(t, sut.IsUninitialized(v1alpha1.ClusterTestSuite{}))
	})

	t.Run("return true when uninitialized set explicitly", func(t *testing.T) {
		// GIVEN
		given := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteUninitialized,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		// WHEN & THEN
		assert.True(t, sut.IsUninitialized(given))
	})

	t.Run("return false when initialized", func(t *testing.T) {
		// GIVEN
		given := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		// WHEN & THEN
		assert.False(t, sut.IsUninitialized(given))
	})
}

func TestIsFinished(t *testing.T) {
	sut := status.Service{}
	t.Run("is not finished when no conditions", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{}
		assert.False(t, sut.IsFinished(givenSuite))
	})

	t.Run("is finished when error", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteError,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		assert.True(t, sut.IsFinished(givenSuite))
	})

	t.Run("is finished when failed", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteFailed,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		assert.True(t, sut.IsFinished(givenSuite))
	})

	t.Run("is finished when succeeded", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteSucceeded,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		assert.True(t, sut.IsFinished(givenSuite))
	})

	t.Run("is not finished when running", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
			},
		}
		assert.False(t, sut.IsFinished(givenSuite))
	})
}

func TestInitialize(t *testing.T) {

	t.Run("when tests not found", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		givenSuite := v1alpha1.ClusterTestSuite{}
		// WHEN
		actualStatus, err := sut.InitializeTests(givenSuite, nil)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actualStatus)
		require.NotNil(t, actualStatus.StartTime)
		assert.Equal(t, actualStatus.StartTime.Time, getStartTime())
		require.NotNil(t, actualStatus.CompletionTime)
		assert.Equal(t, actualStatus.CompletionTime.Time, getStartTime().Add(getTimeInc()))
		assert.Empty(t, actualStatus.Results)
		require.Len(t, actualStatus.Conditions, 1)
		assert.Equal(t, actualStatus.Conditions[0].Type, v1alpha1.SuiteSucceeded)
		assert.Equal(t, actualStatus.Conditions[0].Status, v1alpha1.StatusTrue)
	})

	t.Run("when some tests found", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		givenSuite := v1alpha1.ClusterTestSuite{}
		givenTests := []v1alpha1.TestDefinition{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-1",
					Namespace: "ns-1"},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-2",
					Namespace: "ns-2",
				},
			},
		}
		// WHEN
		actualStatus, err := sut.InitializeTests(givenSuite, givenTests)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actualStatus)
		assert.Equal(t, actualStatus.StartTime.Time, getStartTime())
		assert.Nil(t, actualStatus.CompletionTime)
		require.Len(t, actualStatus.Conditions, 1)
		assert.Equal(t, v1alpha1.SuiteRunning, actualStatus.Conditions[0].Type)
		assert.Equal(t, v1alpha1.StatusTrue, actualStatus.Conditions[0].Status)
		require.Len(t, actualStatus.Results, 2)
		assert.Equal(t, "test-1", actualStatus.Results[0].Name)
		assert.Equal(t, "ns-1", actualStatus.Results[0].Namespace)
		assert.Equal(t, v1alpha1.TestNotYetScheduled, actualStatus.Results[0].Status)
		assert.Equal(t, "test-2", actualStatus.Results[1].Name)
		assert.Equal(t, "ns-2", actualStatus.Results[1].Namespace)
		assert.Equal(t, v1alpha1.TestNotYetScheduled, actualStatus.Results[1].Status)
	})
}

func TestSetSuiteCondition(t *testing.T) {
	sut := status.Service{}
	t.Run("when conditions list is empty, ", func(t *testing.T) {
		stat := &v1alpha1.TestSuiteStatus{}
		sut.SetSuiteCondition(stat, v1alpha1.SuiteRunning, "Reason", "Message")
		require.Len(t, stat.Conditions, 1)
		assert.Equal(t, stat.Conditions[0].Type, v1alpha1.SuiteRunning)
		assert.Equal(t, stat.Conditions[0].Status, v1alpha1.StatusTrue)
		assert.Equal(t, stat.Conditions[0].Reason, "Reason")
		assert.Equal(t, stat.Conditions[0].Message, "Message")

	})

	t.Run("when other conditions were set and add new one", func(t *testing.T) {
		stat := &v1alpha1.TestSuiteStatus{
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:    v1alpha1.SuiteUninitialized,
					Status:  v1alpha1.StatusTrue,
					Message: "old message",
					Reason:  "OldReason",
				},
			},
		}
		sut.SetSuiteCondition(stat, v1alpha1.SuiteRunning, "reason", "message")
		require.Len(t, stat.Conditions, 2)
		assert.Equal(t, stat.Conditions[0].Type, v1alpha1.SuiteUninitialized)
		assert.Equal(t, stat.Conditions[0].Status, v1alpha1.StatusFalse)
		assert.Empty(t, stat.Conditions[0].Reason)
		assert.Empty(t, stat.Conditions[0].Message)

		assert.Equal(t, stat.Conditions[1].Type, v1alpha1.SuiteRunning)
		assert.Equal(t, stat.Conditions[1].Status, v1alpha1.StatusTrue)
		assert.Equal(t, stat.Conditions[1].Reason, "reason")
		assert.Equal(t, stat.Conditions[1].Message, "message")

	})

	t.Run("when updating current condition", func(t *testing.T) {
		stat := &v1alpha1.TestSuiteStatus{
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:   v1alpha1.SuiteRunning,
					Status: v1alpha1.StatusFalse,
				},
			},
		}
		sut.SetSuiteCondition(stat, v1alpha1.SuiteRunning, "reason", "message")
		require.Len(t, stat.Conditions, 1)
		assert.Equal(t, stat.Conditions[0].Type, v1alpha1.SuiteRunning)
		assert.Equal(t, stat.Conditions[0].Status, v1alpha1.StatusTrue)
		assert.Equal(t, stat.Conditions[0].Reason, "reason")
		assert.Equal(t, stat.Conditions[0].Message, "message")

	})
}

func TestEnsureStatusIsUpToDate(t *testing.T) {
	t.Run("when no tests found and suite is already finished", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(nil)
		// WHEN
		suite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteSucceeded,
						Status: v1alpha1.StatusTrue,
					},
				},
				StartTime:      &v1.Time{Time: getStartTime().Add(-2 * time.Hour)},
				CompletionTime: &v1.Time{Time: getStartTime().Add(-time.Hour)},
			},
		}
		noChangesExpected := suite.Status
		stat, err := sut.EnsureStatusIsUpToDate(suite, nil)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, noChangesExpected, *stat)
	})

	t.Run("when pods not yet started", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(nil)
		// WHEN
		suite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-name",
						Namespace: "default",
						Status:    v1alpha1.TestNotYetScheduled,
					},
				},
			},
		}
		noChangesExpected := suite.Status
		stat, err := sut.EnsureStatusIsUpToDate(suite, nil)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, noChangesExpected, *stat)
	})

	t.Run("when first pod is running its phase is updated", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(nil)
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-123",
								StartTime: &v1.Time{Time: getStartTime()},
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getPodAInStatus(
				v12.PodStatus{
					Phase: v12.PodRunning,
				}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:   v1alpha1.SuiteRunning,
					Status: v1alpha1.StatusTrue,
				},
			},
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestRunning,
					Executions: []v1alpha1.TestExecution{
						{
							ID:        "octopus-testing-pod-123",
							PodPhase:  v12.PodRunning,
							StartTime: &v1.Time{Time: getStartTime()},
						},
					},
				},
			},
		}, *stat)
	})

	t.Run("when some pods are running and some are already failed", func(t *testing.T) {
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-123",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
					{
						Name:      "test-b",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-456",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getPodAInStatus(v12.PodStatus{
				Phase: v12.PodRunning,
			}),
			getPodBInStatus(v12.PodStatus{
				Phase:   v12.PodFailed,
				Reason:  "failedReason",
				Message: "failedMessage",
			}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:   v1alpha1.SuiteRunning,
					Status: v1alpha1.StatusTrue,
				},
			},
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestRunning,
					Executions: []v1alpha1.TestExecution{
						{
							ID:        "octopus-testing-pod-123",
							PodPhase:  v12.PodRunning,
							StartTime: &v1.Time{Time: getTimeInPast()},
						},
					},
				},
				{
					Name:      "test-b",
					Namespace: "default",
					Status:    v1alpha1.TestFailed,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             "octopus-testing-pod-456",
							PodPhase:       v12.PodFailed,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime()},
							Reason:         "failedReason",
							Message:        "failedMessage",
						},
					},
				},
			},
		}, *stat)

	})

	t.Run("when all pods finished successfully", func(t *testing.T) {
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-123",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
					{
						Name:      "test-b",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-456",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getPodAInStatus(v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
			getPodBInStatus(v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			CompletionTime: &v1.Time{Time: getStartTime().Add(getTimeInc() * 2)},
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:   v1alpha1.SuiteRunning,
					Status: v1alpha1.StatusFalse,
				},
				{
					Type:   v1alpha1.SuiteSucceeded,
					Status: v1alpha1.StatusTrue,
				},
			},
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestSucceeded,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             "octopus-testing-pod-123",
							PodPhase:       v12.PodSucceeded,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime()},
						},
					},
				},
				{
					Name:      "test-b",
					Namespace: "default",
					Status:    v1alpha1.TestSucceeded,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             "octopus-testing-pod-456",
							PodPhase:       v12.PodSucceeded,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime().Add(getTimeInc())},
						},
					},
				},
			},
		}, *stat)

	})

	t.Run("when all pods finished, one in failed state", func(t *testing.T) {
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteRunning,
						Status: v1alpha1.StatusTrue,
					},
				},
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-123",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
					{
						Name:      "test-b",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        "octopus-testing-pod-456",
								StartTime: &v1.Time{Time: getTimeInPast()},
								PodPhase:  v12.PodRunning,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getPodAInStatus(v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
			getPodBInStatus(v12.PodStatus{
				Phase:   v12.PodFailed,
				Message: "failedMessage",
				Reason:  "failedReason",
			}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			CompletionTime: &v1.Time{Time: getStartTime().Add(getTimeInc() * 2)},
			Conditions: []v1alpha1.TestSuiteCondition{
				{
					Type:   v1alpha1.SuiteRunning,
					Status: v1alpha1.StatusFalse,
				},
				{
					Type:   v1alpha1.SuiteFailed,
					Status: v1alpha1.StatusTrue,
				},
			},
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestSucceeded,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             "octopus-testing-pod-123",
							PodPhase:       v12.PodSucceeded,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime()},
						},
					},
				},
				{
					Name:      "test-b",
					Namespace: "default",
					Status:    v1alpha1.TestFailed,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             "octopus-testing-pod-456",
							PodPhase:       v12.PodFailed,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime().Add(getTimeInc())},
							Message:        "failedMessage",
							Reason:         "failedReason",
						},
					},
				},
			},
		}, *stat)
	})
}

func TestMarkAsScheduled(t *testing.T) {
	// GIVEN
	sut := status.NewService(mockNowProvider())
	// WHEN
	actStatus, err := sut.MarkAsScheduled(v1alpha1.TestSuiteStatus{
		Conditions: []v1alpha1.TestSuiteCondition{
			{
				Type: v1alpha1.SuiteRunning,
			},
		},
		StartTime: &v1.Time{Time: getTimeInPast()},
		Results: []v1alpha1.TestResult{
			{
				Name:      "test-a",
				Namespace: "default",
				Status:    v1alpha1.TestNotYetScheduled,
			},
		},
	}, "test-a", "default", "octopus-testing-pod-123")
	// THEN
	require.NoError(t, err)
	require.Equal(t, v1alpha1.TestSuiteStatus{
		Conditions: []v1alpha1.TestSuiteCondition{
			{
				Type: v1alpha1.SuiteRunning,
			},
		},
		StartTime: &v1.Time{Time: getTimeInPast()},
		Results: []v1alpha1.TestResult{
			{
				Name:      "test-a",
				Namespace: "default",
				Status:    v1alpha1.TestScheduled,
				Executions: []v1alpha1.TestExecution{
					{
						ID:        "octopus-testing-pod-123",
						StartTime: &v1.Time{Time: getStartTime()},
					},
				},
			},
		},
	}, actStatus)
}

func TestGetExecutionsInProgress(t *testing.T) {
	sut := status.NewService(nil)
	t.Run("returns nil if no tests to run", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{}
		// WHEN
		actual := sut.GetExecutionsInProgress(suite)
		// THEN
		require.Empty(t, actual)
	})
	t.Run("returns only Pending and Running executions", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name: "test-1",
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-111",
								PodPhase: v12.PodSucceeded,
							},
							{
								ID:       "id-112",
								PodPhase: v12.PodRunning,
							},
							{
								ID:       "id-113",
								PodPhase: v12.PodFailed,
							},
						},
					},
					{
						Name: "test-2",
						Executions: []v1alpha1.TestExecution{
							{
								ID:       "id-211",
								PodPhase: v12.PodPending,
							},
							{
								ID:       "id-212",
								PodPhase: v12.PodFailed,
							},
						},
					},
				},
			},
		}
		// WHEN
		actual := sut.GetExecutionsInProgress(suite)
		// THEN
		require.Len(t, actual, 2)
		assert.Contains(t, actual, v1alpha1.TestExecution{ID: "id-112", PodPhase: v12.PodRunning})
		assert.Contains(t, actual, v1alpha1.TestExecution{ID: "id-211", PodPhase: v12.PodPending})

	})

}

func mockNowProvider() func() time.Time {
	startTime := getStartTime()
	return func() time.Time {
		defer func() {
			startTime = startTime.Add(getTimeInc())
		}()
		return startTime
	}
}

func getTimeInPast() time.Time {
	return getStartTime().Add(time.Hour * 1)
}
func getStartTime() time.Time {
	return time.Date(2019, 3, 1, 10, 0, 0, 0, time.UTC)
}

func getTimeInc() time.Duration {
	return time.Second
}

func getPodAInStatus(podStatus v12.PodStatus) v12.Pod {
	return v12.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "octopus-testing-pod-123",
			Namespace: "default",
			Labels: map[string]string{
				"testing.kyma-project.io/created-by-octopus": "true",
				"testing.kyma-project.io/suite-name":         "test-all",
				"testing.kyma-project.io/def-name":           "test-a",
			},
		},
		Status: podStatus,
	}
}

func getPodBInStatus(podStatus v12.PodStatus) v12.Pod {
	return v12.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "octopus-testing-pod-456",
			Namespace: "default",
			Labels: map[string]string{
				"testing.kyma-project.io/created-by-octopus": "true",
				"testing.kyma-project.io/suite-name":         "test-all",
				"testing.kyma-project.io/def-name":           "test-b",
			},
		},
		Status: podStatus,
	}
}

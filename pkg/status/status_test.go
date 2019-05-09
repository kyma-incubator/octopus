package status_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)




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
				Conditions: conditionSuiteRunning(),
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
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        getPodNameForTestA(0),
								StartTime: &v1.Time{Time: getStartTime()},
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getTestPodAInStatus(0,
				v12.PodStatus{
					Phase: v12.PodRunning,
				}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			Conditions: conditionSuiteRunning(),
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestRunning,
					Executions: []v1alpha1.TestExecution{
						{
							ID:        getPodNameForTestA(0),
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
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        getPodNameForTestA(0),
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
								ID:        getPodNameForTestB(0),
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
			getTestPodAInStatus(0, v12.PodStatus{
				Phase: v12.PodRunning,
			}),
			getTestPodBInStatus(0, v12.PodStatus{
				Phase:   v12.PodFailed,
				Reason:  "failedReason",
				Message: "failedMessage",
			}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			Conditions: conditionSuiteRunning(),
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestRunning,
					Executions: []v1alpha1.TestExecution{
						{
							ID:        getPodNameForTestA(0),
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
							ID:             getPodNameForTestB(0),
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
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        getPodNameForTestA(0),
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
								ID:        getPodNameForTestB(0),
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
			getTestPodAInStatus(0, v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
			getTestPodBInStatus(0, v12.PodStatus{
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
							ID:             getPodNameForTestA(0),
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
							ID:             getPodNameForTestB(0),
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
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        getPodNameForTestA(0),
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
								ID:        getPodNameForTestB(0),
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
			getTestPodAInStatus(0, v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
			getTestPodBInStatus(0, v12.PodStatus{
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
							ID:             getPodNameForTestA(0),
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
							ID:             getPodNameForTestB(0),
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

	t.Run("suite is running if all already scheduled pods are finished but new one are not created yet", func(t *testing.T) {
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Spec: v1alpha1.TestSuiteSpec{
				Count: 10,
			},
			Status: v1alpha1.TestSuiteStatus{
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:        getPodNameForTestA(0),
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
			getTestPodAInStatus(0, v12.PodStatus{
				Phase: v12.PodSucceeded,
			}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, v1alpha1.TestSuiteStatus{
			Conditions: conditionSuiteRunning(),
			Results: []v1alpha1.TestResult{
				{
					Name:      "test-a",
					Namespace: "default",
					Status:    v1alpha1.TestRunning,
					Executions: []v1alpha1.TestExecution{
						{
							ID:             getPodNameForTestA(0),
							PodPhase:       v12.PodSucceeded,
							StartTime:      &v1.Time{Time: getTimeInPast()},
							CompletionTime: &v1.Time{Time: getStartTime()},
						},
					},
				},
			},
		}, *stat)
	})

	t.Run("maxRetries: suite is succeeded if tests finally pass", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			Spec: specWithRetries(3),
			Status: v1alpha1.TestSuiteStatus{
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       getPodNameForTestA(0),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(1),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(2),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(3),
								PodPhase: v12.PodRunning,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getTestPodAInStatus(0, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(1, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(2, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(3, v12.PodStatus{Phase: v12.PodSucceeded}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, []v1alpha1.TestSuiteCondition{
			{
				Type:   v1alpha1.SuiteRunning,
				Status: v1alpha1.StatusFalse,
			},
			{
				Type:   v1alpha1.SuiteSucceeded,
				Status: v1alpha1.StatusTrue,
			},
		}, stat.Conditions)

	})

	t.Run("maxRetries: suite is running if pods are not finished", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			Spec: specWithRetries(3),
			Status: v1alpha1.TestSuiteStatus{
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       getPodNameForTestA(0),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(1),
								PodPhase: v12.PodPending,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getTestPodAInStatus(0, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(1, v12.PodStatus{Phase: v12.PodRunning}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, []v1alpha1.TestSuiteCondition{
			{
				Type:   v1alpha1.SuiteRunning,
				Status: v1alpha1.StatusTrue,
			},
		}, stat.Conditions)

	})

	t.Run("maxRetries: suite is running if all test failed but maxRetries not reached", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			Spec: specWithRetries(3),
			Status: v1alpha1.TestSuiteStatus{
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       getPodNameForTestA(0),
								PodPhase: v12.PodFailed,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getTestPodAInStatus(0, v12.PodStatus{Phase: v12.PodFailed}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, []v1alpha1.TestSuiteCondition{
			{
				Type:   v1alpha1.SuiteRunning,
				Status: v1alpha1.StatusTrue,
			},
		}, stat.Conditions)

	})

	t.Run("maxRetries: suite is failed if test failed maximum number of times", func(t *testing.T) {
		// GIVEN
		sut := status.NewService(mockNowProvider())
		suite := v1alpha1.ClusterTestSuite{
			Spec: specWithRetries(3),
			Status: v1alpha1.TestSuiteStatus{
				Conditions: conditionSuiteRunning(),
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Status:    v1alpha1.TestRunning,
						Executions: []v1alpha1.TestExecution{
							{
								ID:       getPodNameForTestA(0),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(1),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(2),
								PodPhase: v12.PodFailed,
							},
							{
								ID:       getPodNameForTestA(3),
								PodPhase: v12.PodRunning,
							},
						},
					},
				},
			},
		}
		// WHEN
		stat, err := sut.EnsureStatusIsUpToDate(suite, []v12.Pod{
			getTestPodAInStatus(0, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(1, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(2, v12.PodStatus{Phase: v12.PodFailed}),
			getTestPodAInStatus(3, v12.PodStatus{Phase: v12.PodFailed}),
		})
		// THEN
		require.NoError(t, err)
		require.NotNil(t, stat)
		assert.Equal(t, []v1alpha1.TestSuiteCondition{
			{
				Type:   v1alpha1.SuiteRunning,
				Status: v1alpha1.StatusFalse,
			},
			{
				Type:   v1alpha1.SuiteFailed,
				Status: v1alpha1.StatusTrue,
			},
		}, stat.Conditions)

	})

}

func specWithRetries(retries int64) v1alpha1.TestSuiteSpec {
	return v1alpha1.TestSuiteSpec{
		MaxRetries: retries,
	}
}

func conditionSuiteRunning() []v1alpha1.TestSuiteCondition {
	return []v1alpha1.TestSuiteCondition{
		{
			Type:   v1alpha1.SuiteRunning,
			Status: v1alpha1.StatusTrue,
		},
	}
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

func getPodNameForTestA(exec int) string {
	return fmt.Sprintf("oct-tp-test-all-test-a-%d", exec)
}

func getPodNameForTestB(exec int) string {
	return fmt.Sprintf("oct-tp-test-all-test-b-%d", exec)
}

func getTestPodAInStatus(exec int, podStatus v12.PodStatus) v12.Pod {
	return v12.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      getPodNameForTestA(exec),
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

func getTestPodBInStatus(exec int, podStatus v12.PodStatus) v12.Pod {
	return v12.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      getPodNameForTestB(exec),
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

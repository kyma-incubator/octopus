package v1alpha1_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"fmt"
	"github.com/stretchr/testify/require"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func TestIsUninitialized(t *testing.T) {

	t.Run("return true when suite has empty status", func(t *testing.T) {
		assert.True(t, (&v1alpha1.ClusterTestSuite{}).IsUninitialized())
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
		assert.True(t, given.IsUninitialized())
	})

	for _, tp := range []v1alpha1.TestSuiteConditionType{v1alpha1.SuiteSucceeded, v1alpha1.SuiteRunning, v1alpha1.SuiteFailed} {
		t.Run(fmt.Sprintf("return false when suite is %s", tp), func(t *testing.T) {
			// GIVEN
			given := v1alpha1.ClusterTestSuite{
				Status: v1alpha1.TestSuiteStatus{
					Conditions: []v1alpha1.TestSuiteCondition{
						{
							Type:   v1alpha1.SuiteSucceeded,
							Status: v1alpha1.StatusTrue,
						},
					},
				},
			}
			// WHEN & THEN
			assert.False(t, given.IsUninitialized())
		})
	}

	t.Run("suite is uninitialized if is in error state and error occurred during initialization", func(t *testing.T) {
		// GIVEN
		given := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteError,
						Status: v1alpha1.StatusTrue,
						Reason: v1alpha1.ReasonErrorOnInitialization,
					},
				},
			},
		}
		// WHEN & THEN
		assert.True(t, given.IsUninitialized())
	})

	t.Run("suite is initialized if is in error state and error occurred after initialization", func(t *testing.T) {
		// GIVEN
		given := v1alpha1.ClusterTestSuite{
			Status: v1alpha1.TestSuiteStatus{
				Conditions: []v1alpha1.TestSuiteCondition{
					{
						Type:   v1alpha1.SuiteError,
						Status: v1alpha1.StatusTrue,
						Reason: "other reason",
					},
				},
			},
		}
		// WHEN & THEN
		assert.False(t, given.IsUninitialized())

	})
}

func TestIsFinished(t *testing.T) {
	t.Run("is not finished when no conditions", func(t *testing.T) {
		givenSuite := v1alpha1.ClusterTestSuite{}
		assert.False(t, givenSuite.IsFinished())
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
		assert.True(t, givenSuite.IsFinished())
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
		assert.True(t, givenSuite.IsFinished())
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
		assert.True(t, givenSuite.IsFinished())
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
		assert.False(t, givenSuite.IsFinished())
	})
}

func TestSetSuiteCondition(t *testing.T) {

	t.Run("when conditions list is empty, ", func(t *testing.T) {
		stat := &v1alpha1.TestSuiteStatus{}
		stat.SetSuiteCondition(v1alpha1.SuiteRunning, "Reason", "Message")
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
		stat.SetSuiteCondition(v1alpha1.SuiteRunning, "reason", "message")
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
		stat.SetSuiteCondition(v1alpha1.SuiteRunning, "reason", "message")
		require.Len(t, stat.Conditions, 1)
		assert.Equal(t, stat.Conditions[0].Type, v1alpha1.SuiteRunning)
		assert.Equal(t, stat.Conditions[0].Status, v1alpha1.StatusTrue)
		assert.Equal(t, stat.Conditions[0].Reason, "reason")
		assert.Equal(t, stat.Conditions[0].Message, "message")

	})
}

func TestGetExecutionsInProgress(t *testing.T) {
	t.Run("returns nil if no tests to run", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{}
		// WHEN
		actual := suite.GetExecutionsInProgress()
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
		actual := suite.GetExecutionsInProgress()
		// THEN
		require.Len(t, actual, 2)
		assert.Contains(t, actual, v1alpha1.TestExecution{ID: "id-112", PodPhase: v12.PodRunning})
		assert.Contains(t, actual, v1alpha1.TestExecution{ID: "id-211", PodPhase: v12.PodPending})

	})

}


func TestMarkAsScheduled(t *testing.T) {
	// GIVEN
	ts := &v1alpha1.TestSuiteStatus{
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
	}
	// WHEN
	actStatus, err := ts.MarkAsScheduled("test-a", "default", getPodNameForTestA(0), getStartTime())
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
						ID:        getPodNameForTestA(0),
						StartTime: &v1.Time{Time: getStartTime()},
					},
				},
			},
		},
	}, *actStatus)
}

func getTimeInPast() time.Time {
	return getStartTime().Add(time.Hour * 1)
}
func getStartTime() time.Time {
	return time.Date(2019, 3, 1, 10, 0, 0, 0, time.UTC)
}

func getPodNameForTestA(exec int) string {
	return fmt.Sprintf("oct-tp-test-all-test-a-%d", exec)
}

func getPodNameForTestB(exec int) string {
	return fmt.Sprintf("oct-tp-test-all-test-b-%d", exec)
}
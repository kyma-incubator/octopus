package status_test

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
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
						Type:   v1alpha1.SuiteSucceed,
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
		assert.Equal(t, actualStatus.Conditions[0].Type, v1alpha1.SuiteSucceed)
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
					Namespace: "ns-1",},
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
		assert.Equal(t, "test-2", actualStatus.Results[1].Name)
		assert.Equal(t, "ns-2", actualStatus.Results[1].Namespace)

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
		assert.Empty(t, stat.Conditions[0].Reason, )
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

func mockNowProvider() func() time.Time {
	startTime := getStartTime()
	return func() time.Time {
		defer func() {
			startTime = startTime.Add(getTimeInc())
		}()
		return startTime
	}
}
func getStartTime() time.Time {
	return time.Date(2019, 3, 1, 10, 0, 0, 0, time.UTC)
}

func getTimeInc() time.Duration {
	return time.Second
}

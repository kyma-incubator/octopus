package scheduler

import (
	"testing"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepeatStrategyGetConcurrently(t *testing.T) {
	sut := repeatStrategy{}

	t.Run("return nil when no tests defined", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{}
		// WHEN & THEN
		assert.Nil(t, sut.GetTestToRunConcurrently(suite))
	})

	t.Run("ignore tests with disabled concurrency", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Count: 1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: true,
					},
					{
						Name:                "test2",
						DisabledConcurrency: true,
					},
					{
						Name:                "test3",
						DisabledConcurrency: false,
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunConcurrently(suite)
		// THEN
		require.NotNil(t, actual)
		assert.Equal(t, "test3", actual.Name)
	})

	t.Run("ignore tests that were executed required number of times", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Count: 3,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-111"},
							{ID: "id-222"},
							{ID: "id-333"},
						},
					},
					{
						Name:                "test2",
						DisabledConcurrency: false,
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunConcurrently(suite)
		// THEN
		require.NotNil(t, actual)
		assert.Equal(t, "test2", actual.Name)
	})

	t.Run("return nil if no more tests to run", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Count: 1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-111"},
						},
					},
					{
						Name:                "test-2",
						DisabledConcurrency: false,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-222"},
						},
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunConcurrently(suite)
		// THEN
		require.Nil(t, actual)
	})
}

func TestRepeatStrategyGetSequentially(t *testing.T) {
	sut := repeatStrategy{}

	t.Run("return nil when no tests defined", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{Spec: v1alpha1.TestSuiteSpec{Count: 1}}
		// WHEN & THEN
		assert.Nil(t, sut.GetTestToRunSequentially(suite))
	})

	t.Run("ignore tests with enabled concurrency", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{Count: 1},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: false,
					},
					{
						Name:                "test2",
						DisabledConcurrency: false,
					},
					{
						Name:                "test3",
						DisabledConcurrency: true,
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunSequentially(suite)
		// THEN
		require.NotNil(t, actual)
		assert.Equal(t, "test3", actual.Name)
	})

	t.Run("ignore tests that were executed enough number of times", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Count: 3,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: true,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-111"},
							{ID: "id-222"},
							{ID: "id-333"},
						},
					},
					{
						Name:                "test2",
						DisabledConcurrency: true,
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunSequentially(suite)
		// THEN
		require.NotNil(t, actual)
		assert.Equal(t, "test2", actual.Name)
	})

	t.Run("return nil if no more tests to run", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Count: 1,
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:                "test1",
						DisabledConcurrency: true,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-111"},
						},
					},
					{
						Name:                "test-2",
						DisabledConcurrency: true,
						Executions: []v1alpha1.TestExecution{
							{ID: "id-222"},
						},
					},
				},
			},
		}
		// WHEN
		actual := sut.GetTestToRunSequentially(suite)
		// THEN
		require.Nil(t, actual)
	})
}

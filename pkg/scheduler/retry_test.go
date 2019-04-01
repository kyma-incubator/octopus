package scheduler

import (
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"testing"
)

type retryTestCtx struct {
	testedMethod        func(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult
	DisabledConcurrency bool
	testNamePrefix      string
}

func TestRetryStrategy(t *testing.T) {
	testCases := []retryTestCtx{
		{
			testedMethod:        (&retryStrategy{}).GetTestToRunConcurrently,
			DisabledConcurrency: false,
			testNamePrefix:      "get concurrently",
		},
		{
			testedMethod:        (&retryStrategy{}).GetTestToRunSequentially,
			DisabledConcurrency: true,
			testNamePrefix:      "get sequentially",
		},
	}

	for _, tc := range testCases {

		t.Run(fmt.Sprintf("%s returns nil if no test definitions", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)
		})
		var ignoredTestType string
		if tc.DisabledConcurrency {
			ignoredTestType = "concurrent"
		} else {
			ignoredTestType = "sequential"
		}
		t.Run(fmt.Sprintf("ignores %s tests", ignoredTestType), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(5),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							DisabledConcurrency: !tc.DisabledConcurrency,
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)
		})

		t.Run(fmt.Sprintf("%s runs test with no execution yet", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(5),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.NotNil(t, actual)
			assert.Equal(t, "test-a", actual.Name)
		})

		t.Run(fmt.Sprintf("%s runs test that failed but can be retried", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(3),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
							Executions:          executionsWithPhases(v1.PodFailed, v1.PodFailed, v1.PodFailed),
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.NotNil(t, actual)
			assert.Equal(t, "test-a", actual.Name)

		})

		t.Run(fmt.Sprintf("%s ignore tests that have many executions and one of them succeeded", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(3),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
							Executions:          executionsWithPhases(v1.PodFailed, v1.PodFailed, v1.PodSucceeded),
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)

		})

		t.Run(fmt.Sprintf("%s ignore tests that are currently running", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(3),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
							Executions:          executionsWithPhases(v1.PodRunning),
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)
		})

		t.Run(fmt.Sprintf("%s ignore tests that are currently pending", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(3),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
							Executions:          executionsWithPhases(v1.PodPending),
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)
		})

		t.Run(fmt.Sprintf("%s ignore tests that failed maxRetries times", tc.testNamePrefix), func(t *testing.T) {
			// GIVEN
			suite := v1alpha1.ClusterTestSuite{
				Spec: specWithRetries(3),
				Status: v1alpha1.TestSuiteStatus{
					Results: []v1alpha1.TestResult{
						{
							Name:                "test-a",
							DisabledConcurrency: tc.DisabledConcurrency,
							Executions:          executionsWithPhases(v1.PodFailed, v1.PodFailed, v1.PodFailed, v1.PodFailed),
						},
					},
				},
			}
			// WHEN
			actual := tc.testedMethod(suite)
			// THEN
			require.Nil(t, actual)

		})
	}

}

func specWithRetries(i int) v1alpha1.TestSuiteSpec {
	return v1alpha1.TestSuiteSpec{
		MaxRetries: int64(i),
	}
}

func executionsWithPhases(phases ...v1.PodPhase) []v1alpha1.TestExecution {
	if len(phases) == 0 {
		return nil
	}
	out := make([]v1alpha1.TestExecution, len(phases))
	for idx, phase := range phases {
		out[idx] = v1alpha1.TestExecution{
			PodPhase: phase,
			ID:       fmt.Sprintf("pod-%d", idx+1),
		}
	}
	return out
}

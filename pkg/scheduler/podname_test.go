package scheduler_test

import (
	"k8s.io/apimachinery/pkg/util/rand"
	"strings"
	"testing"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodNameGenerator(t *testing.T) {
	sut := scheduler.PodNameGenerator{}
	t.Run("when first pod to create", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
					},
				},
			},
		}
		// WHEN
		actual, err := sut.GetName(suite, getTestDefinitionA())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "oct-tp-test-all-test-a-0", actual)
	})

	t.Run("when next pod for test to create", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
						Executions: []v1alpha1.TestExecution{
							{
								ID: "oct-tp-test-all-test-a-0",
							},
							{
								ID: "oct-tp-test-all-test-a-1",
							},
						},
					},
				},
			},
		}

		// WHEN
		actual, err := sut.GetName(suite, getTestDefinitionA())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "oct-tp-test-all-test-a-2", actual)
	})

	t.Run("returns error when cannot find test result for test definition", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-all",
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "another namespace",
					},
				},
			},
		}

		// WHEN
		_, err := sut.GetName(suite, getTestDefinitionA())
		// THEN
		require.Error(t, err)
	})

	t.Run("returns error when name is too long", func(t *testing.T) {
		// GIVEN
		suite := v1alpha1.ClusterTestSuite{
			ObjectMeta: v1.ObjectMeta{
				Name: rand.String(300),
			},
			Status: v1alpha1.TestSuiteStatus{
				Results: []v1alpha1.TestResult{
					{
						Name:      "test-a",
						Namespace: "default",
					},
				},
			},
		}

		// WHEN
		_, err := sut.GetName(suite, getTestDefinitionA())
		// THEN
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "generated pod name is too long"))
	})

}

func getTestDefinitionA() v1alpha1.TestDefinition {
	return v1alpha1.TestDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-a",
			Namespace: "default",
		},
	}
}

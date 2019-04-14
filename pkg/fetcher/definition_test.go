package fetcher_test

import (
	"testing"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFindMatching(t *testing.T) {
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)

	t.Run("return all if no selectors specified", func(t *testing.T) {
		// GIVEN
		fakeCli := fake.NewFakeClientWithScheme(sch, &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-def",
				Namespace: "anynamespace",
			},
		})
		service := fetcher.NewForDefinition(fakeCli)

		// WHEN
		out, err := service.FindMatching(v1alpha1.ClusterTestSuite{})
		// THEN
		require.NoError(t, err)
		assert.Len(t, out, 1)
	})

	t.Run("return tests selected by names", func(t *testing.T) {
		// GIVEN
		testA := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-a",
				Namespace: "test-a",
			},
		}
		testB := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-b",
				Namespace: "test-b",
			},
		}
		testC := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-c",
				Namespace: "test-c",
			},
		}

		fakeCli := fake.NewFakeClientWithScheme(sch,
			testA, testB, testC,
		)
		service := fetcher.NewForDefinition(fakeCli)
		// WHEN
		out, err := service.FindMatching(v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Selectors: v1alpha1.TestsSelector{
					MatchNames: []v1alpha1.TestDefReference{
						{
							Name:      "test-a",
							Namespace: "test-a",
						},
						{
							Name:      "test-b",
							Namespace: "test-b",
						},
					},
				},
			},
		})
		// THEN
		require.NoError(t, err)
		assert.Len(t, out, 2)
		assert.Contains(t, out, *testA)
		assert.Contains(t, out, *testB)

	})

	t.Run("return error if test selected by name does not exist", func(t *testing.T) {
		// GIVEN
		fakeCli := fake.NewFakeClientWithScheme(sch)
		service := fetcher.NewForDefinition(fakeCli)
		// WHEN
		_, err := service.FindMatching(v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Selectors: v1alpha1.TestsSelector{
					MatchNames: []v1alpha1.TestDefReference{
						{
							Name:      "test-does-not-exist",
							Namespace: "test-does-not-exist",
						},
					},
				},
			},
		})
		// THEN
		require.EqualError(t, err, "while fetching test definition from selector [name: test-does-not-exist, namespace: test-does-not-exist]: testdefinitions.testing.kyma-project.io \"test-does-not-exist\" not found")
	})

}

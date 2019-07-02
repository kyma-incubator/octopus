package fetcher_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/octopus/pkg/humanerr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
				UID:       "test-uid",
				Name:      "test-a",
				Namespace: "test-a",
			},
		}
		testB := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-b",
				Namespace: "test-b",
			},
		}
		testC := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
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

	t.Run("return tests selected by label expressions", func(t *testing.T) {
		// GIVEN
		testA := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-a",
				Namespace: "test-a",
				Labels: map[string]string{
					"test": "true",
				},
			},
		}
		testB := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-b",
				Namespace: "test-b",
				Labels: map[string]string{
					"test": "false",
				},
			},
		}
		testC := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-c",
				Namespace: "test-c",
				Labels: map[string]string{
					"other": "123",
				},
			},
		}

		fakeCli := fake.NewFakeClientWithScheme(sch,
			testA, testB, testC,
		)
		mockReader := &mockListReader{
			fakeCli: fakeCli,
			listResults: [][]v1alpha1.TestDefinition{
				{*testC},
				{*testA},
			},
		}
		service := fetcher.NewForDefinition(mockReader)
		// WHEN
		out, err := service.FindMatching(v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Selectors: v1alpha1.TestsSelector{
					MatchLabelExpressions: []string{
						"other",
						"test=true",
					},
				},
			},
		})
		// THEN
		require.NoError(t, err)
		assert.Len(t, out, 2)
		assert.Contains(t, out, *testA)
		assert.Contains(t, out, *testC)
	})

	t.Run("return tests returns unique results", func(t *testing.T) {
		// GIVEN
		testA := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-a",
				Namespace: "test-a",
				Labels: map[string]string{
					"test": "true",
				},
			},
		}
		testB := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-b",
				Namespace: "test-b",
				Labels: map[string]string{
					"test": "false",
				},
			},
		}
		testC := &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				UID:       "test-uid",
				Name:      "test-c",
				Namespace: "test-c",
				Labels: map[string]string{
					"other": "123",
				},
			},
		}

		fakeCli := fake.NewFakeClientWithScheme(sch,
			testA, testB, testC,
		)
		mockReader := &mockListReader{
			fakeCli: fakeCli,
			listResults: [][]v1alpha1.TestDefinition{
				{*testA},
			},
		}
		service := fetcher.NewForDefinition(mockReader)
		// WHEN
		out, err := service.FindMatching(v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Selectors: v1alpha1.TestsSelector{
					MatchNames: []v1alpha1.TestDefReference{
						{
							Name:      "test-a",
							Namespace: "test-a",
						},
					},
					MatchLabelExpressions: []string{
						"test=true",
					},
				},
			},
		})
		// THEN
		require.NoError(t, err)
		assert.Len(t, out, 1)
		assert.Contains(t, out, *testA)
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
							Name:      "name",
							Namespace: "ns",
						},
					},
				},
			},
		})
		// THEN
		require.EqualError(t, err, "while fetching test definition from selector [name: name, namespace: ns]: testdefinitions.testing.kyma-project.io \"name\" not found")
		herr, ok := humanerr.GetHumanReadableError(err)
		require.True(t, ok)
		assert.Equal(t, "Test Definition [name: name, namespace: ns] does not exist", herr.Message)
	})

	t.Run("return internal error when fetching selected tests failed", func(t *testing.T) {
		// GIVEN
		errClient := &mockErrReader{err: errors.New("some error")}
		service := fetcher.NewForDefinition(errClient)

		// WHEN
		_, err := service.FindMatching(v1alpha1.ClusterTestSuite{
			Spec: v1alpha1.TestSuiteSpec{
				Selectors: v1alpha1.TestsSelector{
					MatchNames: []v1alpha1.TestDefReference{
						{
							Name:      "name",
							Namespace: "ns",
						},
					},
				},
			},
		})
		// THEN
		require.EqualError(t, err, "while fetching test definition from selector [name: name, namespace: ns]: some error")
		herr, ok := humanerr.GetHumanReadableError(err)
		require.True(t, ok)
		assert.Equal(t, "Internal error", herr.Message)

	})

}

type mockErrReader struct {
	err error
}

func (m *mockErrReader) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return m.err
}

func (m *mockErrReader) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	return m.err
}

type mockListReader struct {
	fakeCli     client.Reader
	listResults [][]v1alpha1.TestDefinition
	calls       uint
}

func (m *mockListReader) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return m.fakeCli.Get(ctx, key, obj)
}

func (m *mockListReader) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	result := m.listResults[m.calls]
	m.calls++
	list.(*v1alpha1.TestDefinitionList).Items = append(list.(*v1alpha1.TestDefinitionList).Items, result...)
	return nil
}

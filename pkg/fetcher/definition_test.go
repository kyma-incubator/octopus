package fetcher_test

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
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
	t.Run("No filters", func(t *testing.T) {
		// GIVEN
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
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

	t.Run("Filter by labels", func(t *testing.T) {
		// GIVEN
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		fakeCli := fake.NewFakeClientWithScheme(sch, &v1alpha1.TestDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-def",
				Namespace: "anynamespace",
			},
		})
		spyCli := &spyReader{underlying: fakeCli}
		service := fetcher.NewForDefinition(spyCli)
		suite := v1alpha1.ClusterTestSuite{Spec: v1alpha1.TestSuiteSpec{
			Selectors: v1alpha1.TestsSelector{
				MatchLabels: []string{"my-label"},
			},
		}}
		// WHEN
		out, err := service.FindMatching(suite)
		// THEN
		require.NoError(t, err)
		assert.Len(t, out, 1)
		selector := spyCli.listCallOpts.LabelSelector
		assert.True(t, selector.Matches(labels.Set{"my-label": "", "other-label": ""}))
		assert.False(t, selector.Matches(labels.Set{"other-label": ""}))
	})
}

type spyReader struct {
	underlying   client.Reader
	listCallOpts *client.ListOptions
}

func (s *spyReader) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return s.underlying.Get(ctx, key, obj)
}

func (s *spyReader) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	s.listCallOpts = opts
	return s.underlying.List(ctx, opts, list)
}

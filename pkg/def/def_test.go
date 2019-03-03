package def_test

import (
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/def"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFindMatching(t *testing.T) {
	// GIVEN
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	fakeCli := fake.NewFakeClientWithScheme(sch, &v1alpha1.TestDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-def",
			Namespace: "anynamespace",
		},
	})
	service := def.NewService(fakeCli)
	// WHEN
	out, err := service.FindMatching(v1alpha1.ClusterTestSuite{})
	// THEN
	require.NoError(t, err)
	assert.Len(t, out, 1)

}

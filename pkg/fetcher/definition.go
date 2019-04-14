package fetcher

import (
	"context"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewForDefinition(r client.Reader) *Definition {
	return &Definition{
		reader: r,
	}
}

type Definition struct {
	reader client.Reader
}

func (s *Definition) FindMatching(suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	ctx := context.TODO()
	if len(suite.Spec.Selectors.MatchNames) > 0 {
		return s.findByNames(ctx, suite)
	}
	// TODO(aszecowka) later so far we return all test definitions for all namespaces (https://github.com/kyma-incubator/octopus/issues/7)
	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(ctx, &client.ListOptions{Namespace: ""}, &list); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}
	return list.Items, nil
}

func (s *Definition) findByNames(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	var list []v1alpha1.TestDefinition
	for _, tRef := range suite.Spec.Selectors.MatchNames {
		def := v1alpha1.TestDefinition{}
		err := s.reader.Get(ctx, types.NamespacedName{Name: tRef.Name, Namespace: tRef.Namespace}, &def)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching test definition from selector [name: %s, namespace: %s]", tRef.Name, tRef.Namespace)
		}
		list = append(list, def)
	}
	return list, nil
}

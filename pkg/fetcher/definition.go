package fetcher

import (
	"context"

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
	// TODO(aszecowka) later so far we return all test definitions for all namespaces (https://github.com/kyma-incubator/octopus/issues/7)
	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(context.TODO(), &client.ListOptions{Namespace: ""}, &list); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}
	return list.Items, nil
}

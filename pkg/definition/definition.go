package definition

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewService(r client.Reader) *Service {
	return &Service{
		reader: r,
	}
}

type Service struct {
	reader client.Reader
}

func (s *Service) FindMatching(suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	// TODO later so far we return all test definitions for all namespaces
	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(context.TODO(), &client.ListOptions{Namespace: ""}, &list); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}
	return list.Items, nil
}

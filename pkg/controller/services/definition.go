package services

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// findTestsThatMatches(suite *testingv1alpha1.TestSuite) ([]testingv1alpha1.Definitions, error)

type Definitions struct {
	reader client.Reader
}

func (s *Definitions) FindMatchingDefinitions(suite *v1alpha1.TestSuite) ([]v1alpha1.TestDefinition, error) {
	if suite.Spec.AllTestsSelector {
		return s.findAll()
	}
	// TODO implement rest
	return nil, nil
}

func (s *Definitions) findAll() ([]v1alpha1.TestDefinition, error) {
	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(context.TODO(), &client.ListOptions{Namespace: ""}, &list); err != nil {
		return nil, err
	}
	// TODO paging?

	return list.Items, nil
}

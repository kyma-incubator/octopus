package fetcher

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

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

	labelsSelector, err := buildLabelsSelector(suite)
	if err != nil {
		return nil, err
	}

	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(context.TODO(), &client.ListOptions{Namespace: "", LabelSelector: labelsSelector}, &list); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}

	return list.Items, nil
}

func buildLabelsSelector(suite v1alpha1.ClusterTestSuite) (labels.Selector, error) {
	selector := labels.NewSelector()
	for _, label := range suite.Spec.Selectors.MatchLabels {
		req, err := labels.NewRequirement(label, selection.Exists, []string{})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*req)
	}
	return selector, nil
}

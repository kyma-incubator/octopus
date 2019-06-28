package fetcher

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/humanerr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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

type uniqueTestDefinitions map[types.UID]v1alpha1.TestDefinition

func (s *Definition) FindMatching(suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	ctx := context.TODO()
	acc := make(uniqueTestDefinitions)

	err := s.findByNames(ctx, suite, acc)
	if err != nil {
		return nil, err
	}

	err = s.findByLabelExpressions(ctx, suite, acc)
	if err != nil {
		return nil, err
	}

	if len(acc) > 0 {
		return acc.getValues(), nil
	}

	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(ctx, &client.ListOptions{Namespace: ""}, &list); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}
	return list.Items, nil
}

func (s *Definition) findByNames(ctx context.Context, suite v1alpha1.ClusterTestSuite, acc uniqueTestDefinitions) error {
	for _, tRef := range suite.Spec.Selectors.MatchNames {
		def := v1alpha1.TestDefinition{}
		err := s.reader.Get(ctx, types.NamespacedName{Name: tRef.Name, Namespace: tRef.Namespace}, &def)
		wrappedErr := errors.Wrapf(err, "while fetching test definition from selector [name: %s, namespace: %s]", tRef.Name, tRef.Namespace)
		switch {
		case err == nil:
		case k8serrors.IsNotFound(err):
			return humanerr.NewError(wrappedErr, fmt.Sprintf("Test Definition [name: %s, namespace: %s] does not exist", tRef.Name, tRef.Namespace))
		default:
			return humanerr.NewError(wrappedErr, "Internal error")
		}
		acc[def.UID] = def
	}
	return nil
}

func (s *Definition) findByLabelExpressions(ctx context.Context, suite v1alpha1.ClusterTestSuite, acc uniqueTestDefinitions) error {
	for _, expr := range suite.Spec.Selectors.MatchLabelExpressions {
		selector, err := labels.Parse(expr)
		if err != nil {
			return errors.Wrapf(err, "while parsing label expression [expression: %s]", expr)
		}
		var list v1alpha1.TestDefinitionList
		if err := s.reader.List(ctx, &client.ListOptions{LabelSelector: selector}, &list); err != nil {
			return errors.Wrapf(err, "while fetching test definition from selector [expression: %s]", expr)
		}
		for _, def := range list.Items {
			acc[def.UID] = def
		}
	}
	return nil
}

func (m uniqueTestDefinitions) getValues() []v1alpha1.TestDefinition {
	var list []v1alpha1.TestDefinition
	for _, def := range m {
		list = append(list, def)
	}
	return list
}

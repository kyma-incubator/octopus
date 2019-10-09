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

func (s *Definition) FindMatching(suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	ctx := context.TODO()

	if suite.HasSelector() {
		return s.findBySelector(ctx, suite)
	}

	return s.findAll(ctx, suite)
}

func (s *Definition) findBySelector(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	byNames, err := s.findByNames(ctx, suite)
	if err != nil {
		return nil, err
	}

	byLabelExpressions, err := s.findByLabelExpressions(ctx, suite)
	if err != nil {
		return nil, err
	}

	return s.unique(byNames, byLabelExpressions), nil
}

func (s *Definition) findByNames(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	result := make([]v1alpha1.TestDefinition, 0)
	for _, tRef := range suite.Spec.Selectors.MatchNames {
		def := v1alpha1.TestDefinition{}
		err := s.reader.Get(ctx, types.NamespacedName{Name: tRef.Name, Namespace: tRef.Namespace}, &def)
		wrappedErr := errors.Wrapf(err, "while fetching test definition from selector [name: %s, namespace: %s]", tRef.Name, tRef.Namespace)
		switch {
		case err == nil:
		case k8serrors.IsNotFound(err):
			return nil, humanerr.NewError(wrappedErr, fmt.Sprintf("Test Definition [name: %s, namespace: %s] does not exist", tRef.Name, tRef.Namespace))
		default:
			return nil, humanerr.NewError(wrappedErr, "Internal error")
		}
		result = append(result, def)
	}
	return result, nil
}

func (s *Definition) findByLabelExpressions(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	result := make([]v1alpha1.TestDefinition, 0)
	for _, expr := range suite.Spec.Selectors.MatchLabelExpressions {
		selector, err := labels.Parse(expr)
		if err != nil {
			return nil, errors.Wrapf(err, "while parsing label expression [expression: %s]", expr)
		}
		var list v1alpha1.TestDefinitionList
		if err := s.reader.List(ctx, &list, &client.ListOptions{LabelSelector: selector}); err != nil {
			return nil, errors.Wrapf(err, "while fetching test definition from selector [expression: %s]", expr)
		}
		result = append(result, list.Items...)
	}
	return result, nil
}

func (s *Definition) unique(slices ...[]v1alpha1.TestDefinition) []v1alpha1.TestDefinition {
	unique := make(map[types.UID]v1alpha1.TestDefinition)
	for _, slice := range slices {
		for _, td := range slice {
			unique[td.UID] = td
		}
	}

	result := make([]v1alpha1.TestDefinition, 0, len(unique))
	for _, td := range unique {
		result = append(result, td)
	}

	return result
}

func (s *Definition) findAll(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1alpha1.TestDefinition, error) {
	var list v1alpha1.TestDefinitionList
	if err := s.reader.List(ctx, &list, &client.ListOptions{Namespace: ""}); err != nil {
		return nil, errors.Wrap(err, "while listing test definitions")
	}
	return list.Items, nil
}

package fetcher

import (
	"context"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewForTestingPod(cli client.Reader) *TestPod {
	return &TestPod{
		cli: cli,
	}
}

type TestPod struct {
	cli client.Reader
}

func (s *TestPod) GetPodsForSuite(ctx context.Context, suite v1alpha1.ClusterTestSuite) ([]v1.Pod, error) {
	var out v1.PodList
	reqCreatedBy, err := labels.NewRequirement(v1alpha1.LabelKeyCreatedByOctopus, selection.Equals, []string{"true"})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating '%s' label requirement", v1alpha1.LabelKeyCreatedByOctopus)
	}
	reqSuiteName, err := labels.NewRequirement(v1alpha1.LabelKeySuiteName, selection.Equals, []string{suite.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating '%s' label requirement", v1alpha1.LabelKeySuiteName)
	}

	if err := s.cli.List(ctx, &client.ListOptions{
		Namespace:     v1.NamespaceAll,
		LabelSelector: labels.NewSelector().Add(*reqCreatedBy, *reqSuiteName),
	}, &out); err != nil {
		return nil, errors.Wrapf(err, "while getting pods for suite [%s]", suite.Name)
	}

	// TODO(aszecowka)(later) deal with pagination
	return out.Items, nil
}

package reporter

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/consts"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewService(cli client.Reader) *Service {
	return &Service{
		cli: cli,
	}
}

type Service struct {
	cli client.Reader
}

func (s *Service) GetPodsForSuite(suite v1alpha1.ClusterTestSuite) ([]v1.Pod, error) {
	var out v1.PodList
	reqCreatedBy, err := labels.NewRequirement(consts.LabelKeyCreatedByOctopus, selection.Equals, []string{"true"})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating '%s' label requirement", consts.LabelKeyCreatedByOctopus)
	}
	reqSuiteName, err := labels.NewRequirement(consts.LabelKeySuiteName, selection.Equals, []string{suite.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating '%s' label requirement", consts.LabelKeySuiteName)
	}

	if err := s.cli.List(context.TODO(), &client.ListOptions{
		Namespace:     "", // from all namespaces
		LabelSelector: labels.NewSelector().Add(*reqCreatedBy, *reqSuiteName),
	}, &out); err != nil {
		return nil, errors.Wrapf(err, "while getting pods for suite [%s]", suite.Name)
	}

	// TODO deal with pagination
	return out.Items, nil
}

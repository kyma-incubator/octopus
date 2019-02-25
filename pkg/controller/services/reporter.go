package services

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDDReporter(reader client.Reader) *DDReporter {
	return &DDReporter{
		reader: reader,
	}
}

type DDReporter struct {
	reader client.Reader
}

func (r *DDReporter) GetPodsForSuite(suite *v1alpha1.TestSuite) ([]v1.Pod, error) {
	var pods v1.PodList
	req, err := labels.NewRequirement("octopus.suite.name", selection.Equals, []string{suite.Name})
	if err != nil {
		return nil, err
	}
	if err := r.reader.List(context.TODO(), &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req),
	}, &pods); err != nil {
		return nil, err
	}
	return pods.Items, nil
}

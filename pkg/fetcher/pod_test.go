package fetcher_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/fetcher"
	"github.com/kyma-incubator/octopus/pkg/fetcher/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetPodsForSuite(t *testing.T) {
	// GIVEN
	givenSuite := v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "test-all-suite",
	}}

	givenPod := v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      fmt.Sprintf("%s-my-test-%d", givenSuite.Name, 1),
			Namespace: "aaa",
			Labels: map[string]string{
				v1alpha1.LabelKeyCreatedByOctopus: "true",
				v1alpha1.LabelKeySuiteName:        "test-all-suite",
			},
		},
	}

	mockReader := &automock.Reader{}
	defer mockReader.AssertExpectations(t)

	listOptionMatcher := mock.MatchedBy(func(listOptions *client.ListOptions) bool {
		if listOptions.Namespace != "" {
			return false
		}
		if !listOptions.LabelSelector.Matches(labels.Set(givenPod.Labels)) {
			return false
		}
		return true
	})

	mockReader.On("List", mock.Anything, listOptionMatcher, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			podList, ok := args.Get(2).(*v1.PodList)
			if ok {
				podList.Items = []v1.Pod{givenPod}
			}
		})
	sut := fetcher.NewForTestingPod(mockReader)
	// WHEN
	actualPods, err := sut.GetPodsForSuite(context.TODO(), givenSuite)
	// THEN
	require.NoError(t, err)
	require.Len(t, actualPods, 1)
	assert.Equal(t, givenPod, actualPods[0])

}

func TestGetPodsForSuiteOnError(t *testing.T) {
	givenSuite := v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "test-all-suite",
	}}

	mockReader := &automock.Reader{}
	mockReader.On("List", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error"))
	defer mockReader.AssertExpectations(t)
	sut := fetcher.NewForTestingPod(mockReader)
	// WHEN
	_, err := sut.GetPodsForSuite(context.TODO(), givenSuite)
	// THEN
	require.EqualError(t, err, "while getting pods for suite [test-all-suite]: some error")
}

package reporter_test

import (
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/consts"
	"github.com/kyma-incubator/octopus/pkg/reporter"
	"github.com/kyma-incubator/octopus/pkg/reporter/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestGetPodsForSuite(t *testing.T) {
	// GIVEN
	pod1 := v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      fmt.Sprintf("%s%d", consts.TestingPodGeneratedName, 1),
			Namespace: "aaa",
			Labels: map[string]string{
				consts.LabelKeyCreatedByOctopus: "true",
				consts.LabelKeySuiteName:        "test-all-suite",
			},
		},
	}

	givenSuite := v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "test-all-suite",
	}}

	mockReader := &automock.Reader{}
	defer mockReader.AssertExpectations(t)

	listOptionMatcher := mock.MatchedBy(func(listOptions *client.ListOptions) bool {
		if listOptions.Namespace != "" {
			return false
		}
		if !listOptions.LabelSelector.Matches(labels.Set(pod1.Labels)) {
			return false
		}
		return true
	})

	mockReader.On("List", mock.Anything, listOptionMatcher, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			podList, ok := args.Get(2).(*v1.PodList)
			if ok {
				podList.Items = []v1.Pod{pod1}
			}
		})
	sut := reporter.NewService(mockReader)
	// WHEN
	actualPods, err := sut.GetPodsForSuite(givenSuite)
	// THEN
	require.NoError(t, err)
	require.Len(t, actualPods, 1)
	assert.Equal(t, pod1, actualPods[0])

}

func TestGetPodsForSuiteOnError(t *testing.T) {
	givenSuite := v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "test-all-suite",
	}}

	mockReader := &automock.Reader{}
	mockReader.On("List", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error"))
	defer mockReader.AssertExpectations(t)
	sut := reporter.NewService(mockReader)
	// WHEN
	_, err := sut.GetPodsForSuite(givenSuite)
	// THEN
	require.EqualError(t, err, "while getting pods for suite [test-all-suite]: some error")
}

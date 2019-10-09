package fetcher_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/octopus/pkg/scheduler"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/fetcher"
	"github.com/kyma-incubator/octopus/pkg/fetcher/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetPodsForSuite(t *testing.T) {
	// GIVEN
	givenPod := v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      fmt.Sprintf("%s%d", scheduler.TestingPodPrefix, 1),
			Namespace: "aaa",
			Labels: map[string]string{
				v1alpha1.LabelKeyCreatedByOctopus: "true",
				v1alpha1.LabelKeySuiteName:        "test-all-suite",
			},
		},
	}

	givenSuite := v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "test-all-suite",
	}}

	mockReader := &automock.Reader{}
	defer mockReader.AssertExpectations(t)

	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, v1.AddToScheme(sch))

	cli := fake.NewFakeClientWithScheme(sch, &givenPod)

	sut := fetcher.NewForTestingPod(cli)
	// WHEN
	actualPods, err := sut.GetPodsForSuite(context.TODO(), givenSuite)
	// THEN
	require.NoError(t, err)
	require.Len(t, actualPods, 1)
	assert.Equal(t, givenPod, actualPods[0])

	// WHEN
	actualPods, err = sut.GetPodsForSuite(context.Background(), v1alpha1.ClusterTestSuite{ObjectMeta: v12.ObjectMeta{
		Name: "wrong-name",
	}})
	// THEN
	require.NoError(t, err)
	require.Len(t, actualPods, 0)
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

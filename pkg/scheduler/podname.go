package scheduler

import (
	"fmt"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
)

const (
	TestingPodPrefix = "oct-tp"
)

type PodNameGenerator struct{}

func (P *PodNameGenerator) GetName(suite v1alpha1.ClusterTestSuite, def v1alpha1.TestDefinition) (string, error) {
	// max length of k8s name is 253 characters
	// TODO: add validation
	idx := -1
	for _, tr := range suite.Status.Results {
		if tr.Name == def.Name && tr.Namespace == def.Namespace {
			idx = len(tr.Executions)
			break
		}
	}
	if idx == -1 {
		return "", fmt.Errorf("while generating Pod name for suite [%s] and test definition [name: %s, namespace: %s]", suite.Name, def.Name, def.Namespace)
	}
	return fmt.Sprintf("%s-%s-%s-%d", TestingPodPrefix, suite.Name, def.Name, idx), nil
}

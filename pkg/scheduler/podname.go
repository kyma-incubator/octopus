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
	// TODO (aszecowka): add validation https://github.com/kyma-incubator/octopus/issues/11
	idx := -1
	for _, tr := range suite.Status.Results {
		if tr.Name == def.Name && tr.Namespace == def.Namespace {
			idx = len(tr.Executions)
			break
		}
	}
	if idx == -1 {
		return "", fmt.Errorf("while generating Pod name for suite [%s] and test definition [name: %s, namespace: %s]: the suite has uninitialized status", suite.Name, def.Name, def.Namespace)
	}
	name := fmt.Sprintf("%s-%s-%s-%d", TestingPodPrefix, suite.Name, def.Name, idx)
	if len(name) > 253 {
		return "", fmt.Errorf("generated pod name is too long: [%s]", name)
	}
	return name, nil
}

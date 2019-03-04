package scheduler

import (
	"context"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/consts"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockery -name=StatusProvider -output=automock -outpkg=automock -case=underscore
type StatusProvider interface {
	GetNextToSchedule(suite v1alpha1.ClusterTestSuite) (*v1alpha1.TestResult, error)
	MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName, testNs, podName string) (v1alpha1.TestSuiteStatus, error)
}

func NewService(statusProvider StatusProvider, reader client.Reader, writer client.Writer) *Service {
	return &Service{
		statusProvider: statusProvider,
		reader:         reader,
		writer:         writer,
	}
}

type Service struct {
	statusProvider StatusProvider
	reader         client.Reader
	writer         client.Writer
}

func (s *Service) TrySchedule(suite v1alpha1.ClusterTestSuite) (*v1.Pod, *v1alpha1.TestSuiteStatus, error) {
	tr, err := s.statusProvider.GetNextToSchedule(suite)
	if err != nil {
		return nil, nil, err
	}
	if tr == nil {
		return nil, nil, nil
	}
	def, err := s.getDefinition(tr.Name, tr.Namespace)
	if err != nil {
		return nil, nil, err
	}
	pod, err := s.startPod(suite, def)
	if err != nil {
		return nil, nil, err
	}

	curr, err := s.statusProvider.MarkAsScheduled(suite.Status, tr.Name, tr.Namespace, pod.Name)
	if err != nil {
		return nil, nil, err
	}
	return pod, &curr, nil
}

func (s *Service) getDefinition(name, ns string) (v1alpha1.TestDefinition, error) {
	var out v1alpha1.TestDefinition
	err := s.reader.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, &out)
	if err != nil {
		return v1alpha1.TestDefinition{}, errors.Wrapf(err, "while getting test definition [name: %s, namespace: %s]", name, ns)
	}
	return out, nil
}

func (s *Service) startPod(suite v1alpha1.ClusterTestSuite, def v1alpha1.TestDefinition) (*v1.Pod, error) {
	p := &v1.Pod{}

	p.Spec = def.Spec.Template.Spec
	p.Labels = def.Spec.Template.Labels
	p.Annotations = def.Spec.Template.Annotations

	p.GenerateName = consts.TestingPodGeneratedName
	p.Namespace = def.Namespace

	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	p.Labels[consts.LabelKeySuiteName] = suite.Name
	p.Labels[consts.LabelKeyTestDefName] = def.Name
	p.Labels[consts.LabelKeyCreatedByOctopus] = "true"
	p.Spec.RestartPolicy = v1.RestartPolicyNever

	err := s.writer.Create(context.TODO(), p)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating testing pod for suite [%s] and test definition [name: %s, namespace: %s]", suite.Name, def.Name, def.Namespace)
	}
	return p, nil
}

package scheduler

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//go:generate go run ./../../vendor/github.com/vektra/mockery/cmd/mockery/mockery.go -name=StatusProvider -output=automock -outpkg=automock -case=underscore
type StatusProvider interface {
	GetNextToSchedule(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult
	MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName, testNs, podName string) (v1alpha1.TestSuiteStatus, error)
}

func NewService(statusProvider StatusProvider, reader client.Reader, writer client.Writer, scheme *runtime.Scheme) *Service {
	return &Service{
		statusProvider: statusProvider,
		reader:         reader,
		writer:         writer,
		scheme:         scheme,
	}
}

type Service struct {
	statusProvider StatusProvider
	reader         client.Reader
	writer         client.Writer
	scheme         *runtime.Scheme
}

func (s *Service) TrySchedule(suite v1alpha1.ClusterTestSuite) (*v1.Pod, *v1alpha1.TestSuiteStatus, error) {
	tr := s.statusProvider.GetNextToSchedule(suite)
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
		return nil, nil, errors.Wrapf(err, "while scheduling suite [%s]", suite.Name)
	}
	return pod, &curr, nil
}

func (s *Service) getDefinition(name, ns string) (v1alpha1.TestDefinition, error) {
	var out v1alpha1.TestDefinition
	err := s.reader.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, &out)
	if err != nil {
		return v1alpha1.TestDefinition{}, errors.Wrapf(err, "while getting test fetcher [name: %s, namespace: %s]", name, ns)
	}
	return out, nil
}

func (s *Service) startPod(suite v1alpha1.ClusterTestSuite, def v1alpha1.TestDefinition) (*v1.Pod, error) {
	p := &v1.Pod{}
	// TODO (aszeowka)(later) https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/controller_utils.go#L517-L522
	p.Spec = def.Spec.Template.Spec
	p.Labels = def.Spec.Template.Labels
	p.Annotations = def.Spec.Template.Annotations

	p.GenerateName = fmt.Sprintf("%s-%s-", suite.Name, def.Name)
	p.Namespace = def.Namespace

	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	p.Labels[v1alpha1.LabelKeySuiteName] = suite.Name
	p.Labels[v1alpha1.LabelKeyTestDefName] = def.Name
	p.Labels[v1alpha1.LabelKeyCreatedByOctopus] = "true"
	p.Spec.RestartPolicy = v1.RestartPolicyNever

	if err := controllerutil.SetControllerReference(&suite, p, s.scheme); err != nil {
		return nil, errors.Wrapf(err, "while setting controller reference, suite [%s], pod [name %s, namespace: %s]", suite.Name, p.Name, p.Namespace)
	}

	err := s.writer.Create(context.TODO(), p)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating testing pod for suite [%s] and test fetcher [name: %s, namespace: %s]", suite.Name, def.Name, def.Namespace)
	}
	return p, nil
}

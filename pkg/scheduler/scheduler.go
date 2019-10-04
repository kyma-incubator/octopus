package scheduler

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type StatusProvider interface {
	MarkAsScheduled(status v1alpha1.TestSuiteStatus, testName, testNs, podName string) (v1alpha1.TestSuiteStatus, error)
	GetExecutionsInProgress(suite v1alpha1.ClusterTestSuite) []v1alpha1.TestExecution
}

// nextTestSelectorStrategy
type nextTestSelectorStrategy interface {
	GetTestToRunConcurrently(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult
	GetTestToRunSequentially(suite v1alpha1.ClusterTestSuite) *v1alpha1.TestResult
}

type podNameProvider interface {
	GetName(suite v1alpha1.ClusterTestSuite, def v1alpha1.TestDefinition) (string, error)
}

func NewService(statusProvider StatusProvider, reader client.Reader, writer client.Writer, scheme *runtime.Scheme, logger logr.Logger) *Service {
	return &Service{
		statusProvider: statusProvider,
		reader:         reader,
		writer:         writer,
		scheme:         scheme,
		log:            logger,
	}
}

type Service struct {
	statusProvider StatusProvider
	reader         client.Reader
	writer         client.Writer
	scheme         *runtime.Scheme
	log            logr.Logger
}

func (s *Service) TrySchedule(suite v1alpha1.ClusterTestSuite) (*v1.Pod, *v1alpha1.TestSuiteStatus, error) {
	tr, err := s.getNextToSchedule(suite)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while getting next to schedule")
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
		return nil, nil, errors.Wrapf(err, "while marking suite [%s] as Scheduled", suite.Name)
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

func (s *Service) getNextToSchedule(suite v1alpha1.ClusterTestSuite) (*v1alpha1.TestResult, error) {
	suite = s.normalizeSuite(suite)
	running := s.statusProvider.GetExecutionsInProgress(suite)

	logSuite := s.log.WithValues("suite", suite.Name)
	if len(running) >= int(suite.Spec.Concurrency) {
		logSuite.Info("Cannot get next test to schedule, max concurrency reached", "running", len(running), "concurrency", suite.Spec.Concurrency)
		return nil, nil
	}

	strategy := s.getStrategyForSuite(suite)
	if strategy == nil {
		err := fmt.Errorf("cannot find test selector strategy that is applicable for suite [%s]", suite.Name)
		logSuite.Error(err, "No applicable strategy")
		return nil, err
	}

	if toRunCandidate := strategy.GetTestToRunConcurrently(suite); toRunCandidate != nil {
		return toRunCandidate, nil
	}

	if len(running) == 0 {
		if toRunCandidate := strategy.GetTestToRunSequentially(suite); toRunCandidate != nil {
			return toRunCandidate, nil
		}
	}

	logSuite.Info("No tests to execute right now")
	return nil, nil
}

// TODO this is only workaround, proper implementation will be done here: https://github.com/kyma-incubator/octopus/issues/11
func (s *Service) normalizeSuite(suite v1alpha1.ClusterTestSuite) v1alpha1.ClusterTestSuite {
	if suite.Spec.Concurrency == 0 {
		suite.Spec.Concurrency = 1
	}
	if suite.Spec.Count == 0 {
		suite.Spec.Count = 1
	}
	return suite
}

func (s *Service) getStrategyForSuite(suite v1alpha1.ClusterTestSuite) nextTestSelectorStrategy {
	if suite.Spec.MaxRetries == 0 {
		return &repeatStrategy{}
	} else {
		return &retryStrategy{}
	}

}

func (s *Service) getNameProvider() podNameProvider {
	return &PodNameGenerator{}
}

func (s *Service) startPod(suite v1alpha1.ClusterTestSuite, def v1alpha1.TestDefinition) (*v1.Pod, error) {
	p := &v1.Pod{}
	// TODO (aszeowka)(later) https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/controller_utils.go#L517-L522
	p.Spec = def.Spec.Template.Spec
	p.Labels = def.Spec.Template.Labels
	p.Annotations = def.Spec.Template.Annotations

	name, err := s.getNameProvider().GetName(suite, def)
	if err != nil {
		return nil, err
	}
	p.Name = name
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

	err = s.writer.Create(context.TODO(), p)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating testing pod for suite [%s] and test definition [name: %s, namespace: %s]", suite.Name, def.Name, def.Namespace)
	}
	return p, nil
}

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testsuite

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	testingv1alpha1 "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-incubator/octopus/pkg/fetcher"
	"github.com/kyma-incubator/octopus/pkg/scheduler"
	"github.com/kyma-incubator/octopus/pkg/status"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new ClusterTestSuite Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	statusSvc := status.NewService(time.Now)
	schedulerSvc := scheduler.NewService(statusSvc, mgr.GetClient(), mgr.GetClient(), mgr.GetScheme())
	podSvc := fetcher.NewForTestingPod(mgr.GetClient())
	return &ReconcileTestSuite{
		Client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		scheduler:         schedulerSvc,
		statusService:     statusSvc,
		definitionService: fetcher.NewForDefinition(mgr.GetClient()),
		podSvc:            podSvc,
		log:               logf.Log.WithName("cts_controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller

	c, err := controller.New("testsuite-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterTestSuite
	err = c.Watch(&source.Kind{Type: &testingv1alpha1.ClusterTestSuite{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to Pods
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &testingv1alpha1.ClusterTestSuite{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTestSuite{}

// ReconcileTestSuite reconciles a ClusterTestSuite object
type ReconcileTestSuite struct {
	client.Client
	scheme            *runtime.Scheme
	scheduler         TestScheduler
	podSvc            TestReporter
	statusService     SuiteStatusService
	definitionService TestDefinitionService
	log               logr.Logger
}

const (
	defaultRequeueAfter = time.Second
)

// Reconcile reads that state of the cluster for a ClusterTestSuite object and makes changes based on the state read
// and what is in the ClusterTestSuite.Spec

// Automatically generate RBAC rules to allow the Controller to read and write Pods
// +kubebuilder:rbac:groups=apps,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=testing.kyma-project.io,resources=testsuites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=testing.kyma-project.io,resources=testsuites/status,verbs=get;update;patch
func (r *ReconcileTestSuite) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	// Fetch the ClusterTestSuite suiteCopy
	suite := &testingv1alpha1.ClusterTestSuite{}
	err := r.Get(ctx, request.NamespacedName, suite)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	suiteCopy := suite.DeepCopy()
	logSuite := r.log.WithValues("suite", suite.Name)

	if r.statusService.IsUninitialized(*suiteCopy) {
		logSuite.Info("Initialize suite")
		testDefs, err := r.definitionService.FindMatching(*suiteCopy)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "while looking for matching test definitions for suite [%s]", suiteCopy.Name)
		}
		currStatus, err := r.statusService.InitializeTests(*suiteCopy, testDefs)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "while initializing tests for suite [%s]", suiteCopy.Name)
		}
		suiteCopy.Status = *currStatus
		if err := r.Client.Status().Update(ctx, suiteCopy); err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "while updating status of initialized suite [%s]", suiteCopy.Name)
		}
		return reconcile.Result{Requeue: true}, nil
	}
	if r.statusService.IsFinished(*suiteCopy) {
		logSuite.Info("Do nothing, suite is finished")
		return reconcile.Result{}, nil
	}

	logSuite.Info("Ensuring status is up-to-date")
	updatedStatus, err := r.ensureStatusIsUpToDate(ctx, *suiteCopy)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while ensuring status is up-to-date for suite [%s]", suiteCopy.Name)
	}
	suiteCopy.Status = *updatedStatus
	pod, updatedStatus, err := r.scheduler.TrySchedule(*suiteCopy)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while scheduling next testing pod for suite [%s]", suiteCopy.Name)
	}
	if pod != nil {
		logSuite.Info("Testing pod created", "podName", pod.Name, "podNs", pod.Namespace)
		suiteCopy.Status = *updatedStatus
	}

	if err := r.Client.Status().Update(ctx, suiteCopy); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while updating status of running suite [%s]", suiteCopy.Name)
	}

	if pod != nil {
		// requeue immediately to try schedule other tests
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{Requeue: true, RequeueAfter: defaultRequeueAfter}, nil
}

func (r *ReconcileTestSuite) ensureStatusIsUpToDate(ctx context.Context, suite testingv1alpha1.ClusterTestSuite) (*testingv1alpha1.TestSuiteStatus, error) {
	pods, err := r.podSvc.GetPodsForSuite(ctx, suite)
	if err != nil {
		return nil, err
	}

	return r.statusService.EnsureStatusIsUpToDate(suite, pods)
}

// dependencies
type TestScheduler interface {
	TrySchedule(suite testingv1alpha1.ClusterTestSuite) (*corev1.Pod, *testingv1alpha1.TestSuiteStatus, error)
}

type TestReporter interface {
	GetPodsForSuite(ctx context.Context, suite testingv1alpha1.ClusterTestSuite) ([]corev1.Pod, error)
}

type SuiteStatusService interface {
	EnsureStatusIsUpToDate(suite testingv1alpha1.ClusterTestSuite, pods []corev1.Pod) (*testingv1alpha1.TestSuiteStatus, error)
	InitializeTests(suite testingv1alpha1.ClusterTestSuite, defs []testingv1alpha1.TestDefinition) (*testingv1alpha1.TestSuiteStatus, error)
	IsUninitialized(suite testingv1alpha1.ClusterTestSuite) bool
	IsFinished(suite testingv1alpha1.ClusterTestSuite) bool
}

type TestDefinitionService interface {
	FindMatching(suite testingv1alpha1.ClusterTestSuite) ([]testingv1alpha1.TestDefinition, error)
}

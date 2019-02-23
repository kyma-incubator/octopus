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
	testingv1alpha1 "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TestSuite Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTestSuite{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("testsuite-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to TestSuite
	err = c.Watch(&source.Kind{Type: &testingv1alpha1.TestSuite{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by TestSuite - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &testingv1alpha1.TestSuite{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTestSuite{}

// ReconcileTestSuite reconciles a TestSuite object
type ReconcileTestSuite struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TestSuite object and makes changes based on the state read
// and what is in the TestSuite.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=testing.kyma-project.io,resources=testsuites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=testing.kyma-project.io,resources=testsuites/status,verbs=get;update;patch
func (r *ReconcileTestSuite) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the TestSuite suite
	ctx := context.TODO()
	suite := &testingv1alpha1.TestSuite{}
	err := r.Get(context.TODO(), request.NamespacedName, suite)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	suite = suite.DeepCopy()

	if r.isUninitialized(suite) {
		testDefs, err := r.findTestsThatMatches(suite)
		if err != nil {
			return reconcile.Result{}, nil
		}
		if err := r.initializeTests(suite, testDefs); err != nil {
			return reconcile.Result{}, err
		}

		if err := r.Client.Status().Update(ctx, suite); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	}

	if r.isFinished(suite) {
		return reconcile.Result{}, nil
	}

	// Test Suite is in progress
	if err := r.ensureStatusUpToDate(suite); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.tryScheduleTests(suite); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.Client.Status().Update(ctx, suite); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTestSuite) isUninitialized(suite *testingv1alpha1.TestSuite) bool {
	return true
}

func (r *ReconcileTestSuite) isFinished(suite *testingv1alpha1.TestSuite) bool {
	return false
}

func (r *ReconcileTestSuite) initializeTests(suite *testingv1alpha1.TestSuite, defs []testingv1alpha1.TestDefinition) error {
	return nil
}

// find test definitions
func (r *ReconcileTestSuite) findTestsThatMatches(suite *testingv1alpha1.TestSuite) ([]testingv1alpha1.TestDefinition, error) {
	return nil, nil
}

// get info from pods
func (r *ReconcileTestSuite) ensureStatusUpToDate(suite *testingv1alpha1.TestSuite) error {
	return nil
}

// create Pod
func (r *ReconcileTestSuite) tryScheduleTests(suite *testingv1alpha1.TestSuite) error {
	return nil
}

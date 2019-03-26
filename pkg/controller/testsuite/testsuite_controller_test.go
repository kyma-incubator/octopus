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
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	testingv1alpha1 "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const succeededSuiteTimeout = time.Second * 20

func TestReconcileClusterTestSuite(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	logf.SetLogger(logf.ZapLogger(false))

	suite := &testingv1alpha1.ClusterTestSuite{
		ObjectMeta: metav1.ObjectMeta{Name: "suite-test-ls-command"},
		Spec: testingv1alpha1.TestSuiteSpec{
			Concurrency: 1,
			Count:       2,
		},
	}

	// Setup the Manager and Controller
	mgr, err := manager.New(cfg, manager.Options{})
	require.NoError(t, err)
	c := mgr.GetClient()

	require.NoError(t, add(mgr, newReconciler(mgr)))
	stopMgr, mgrStopped := StartTestManager(t, mgr)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	err = startMockPodController(mgr)
	require.NoError(t, err)

	// WHEN
	// Create the TestDefinition
	err = c.Create(ctx, getTestDefLs())
	require.NoError(t, err)
	defer c.Delete(ctx, getTestDefLs())

	err = c.Create(ctx, getTestDefPwd())
	require.NoError(t, err)
	defer c.Delete(ctx, getTestDefPwd())

	// Create the ClusterTestSuite object and expect the Reconcile
	err = c.Create(ctx, suite)
	require.NoError(t, err)
	defer c.Delete(ctx, suite)

	// THEN
	repeat.FuncAtMost(t, func() error {
		var actualSuite testingv1alpha1.ClusterTestSuite
		err := c.Get(ctx, types.NamespacedName{Name: "suite-test-ls-command"}, &actualSuite)
		//defer func() {
		//	spew.Dump("actualSuite",actualSuite)
		//}()

		if err != nil {
			return err
		}
		succeeded := false
		for _, cond := range actualSuite.Status.Conditions {
			if cond.Type == testingv1alpha1.SuiteSucceeded && cond.Status == testingv1alpha1.StatusTrue {
				succeeded = true
			} else if cond.Status == testingv1alpha1.StatusTrue {
				return fmt.Errorf("suite is in invalid state [%s]", cond.Type)
			}
		}

		if !succeeded {
			return errors.New("suite should be succeeded")
		}

		return nil
	}, succeededSuiteTimeout)

}

// We need to manually change status of Pod
type mockPodReconciler struct {
	cli client.Client
}

func (r *mockPodReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	pod := &v1.Pod{}
	ctx := context.Background()
	err := r.cli.Get(ctx, request.NamespacedName, pod)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.getLogger().Info("got pod", "podName", pod.Name, "podPhase", pod.Status.Phase)
	pod = pod.DeepCopy()
	if pod.Status.Phase == v1.PodPending {
		pod.Status.Phase = v1.PodRunning
		r.getLogger().Info("starting pod", "podName", pod.Name)
		err := r.cli.Status().Update(ctx, pod)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	if pod.Status.Phase == v1.PodRunning {
		pod.Status.Phase = v1.PodSucceeded
		r.getLogger().Info("pod succeeded", "podName", pod.Name)
		err := r.cli.Status().Update(ctx, pod)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *mockPodReconciler) getLogger() logr.Logger {
	return logf.Log.WithName("mock pod controller")
}

func startMockPodController(mgr manager.Manager) error {
	pr := &mockPodReconciler{
		cli: mgr.GetClient(),
	}
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: pr})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func getTestDefLs() *testingv1alpha1.TestDefinition {
	return &testingv1alpha1.TestDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ls-command", Namespace: "default"},
		Spec: testingv1alpha1.TestDefinitionSpec{
			Template: v1.PodTemplateSpec{

				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "test",
							Image:   "alpine:3.9",
							Command: []string{"ls"}},
					},
				},
			},
		},
	}
}

func getTestDefPwd() *testingv1alpha1.TestDefinition {
	return &testingv1alpha1.TestDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pwd-command", Namespace: "default"},
		Spec: testingv1alpha1.TestDefinitionSpec{
			DisableConcurrency: true,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "test",
							Image:   "alpine:3.9",
							Command: []string{"pwd"}},
					},
				},
			},
		},
	}
}

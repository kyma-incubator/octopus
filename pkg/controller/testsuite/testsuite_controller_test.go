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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"testing"
	"time"

	testingv1alpha1 "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(false))

	g := gomega.NewGomegaWithT(t)
	testDef := &testingv1alpha1.TestDefinition{
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
	suite := &testingv1alpha1.ClusterTestSuite{ObjectMeta: metav1.ObjectMeta{Name: "suite-test-ls-command"}}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the TestDefinition
	err = c.Create(context.TODO(), testDef)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), testDef)

	// Create the ClusterTestSuite object and expect the Reconcile
	err = c.Create(context.TODO(), suite)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), suite)

	err = startMockPodController(mgr)
	require.NoError(t, err)

	go func() {
		for {
			// I don't care how many requests will be sent to my Reconciler, but I have to read all of them to not block
			log.Info("Got reconcile request", "req", <-requests)
		}
	}()

	g.Eventually(func() error {
		var actualSuite testingv1alpha1.ClusterTestSuite
		err := c.Get(context.TODO(), types.NamespacedName{Name: "suite-test-ls-command"}, &actualSuite)
		if err != nil {
			return err
		}
		if len(actualSuite.Status.Conditions) != 2 {
			return errors.New("Should have 2 conditions")
		}
		suiteRunning := actualSuite.Status.Conditions[0].Type == testingv1alpha1.SuiteRunning && actualSuite.Status.Conditions[0].Status == testingv1alpha1.StatusTrue
		if suiteRunning {
			return errors.New("suite should not be running")
		}
		suiteSucceeded := actualSuite.Status.Conditions[1].Type == testingv1alpha1.SuiteSucceeded && actualSuite.Status.Conditions[1].Status == testingv1alpha1.StatusTrue
		if !suiteSucceeded {
			return errors.New("suite should be succeeded")
		}

		return nil
	}, timeout).
		Should(gomega.Succeed())

}

// We need to manually change status of Pod
type mockPodReconciler struct {
	cli client.Client
}

func (r *mockPodReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	pod := &v1.Pod{}
	err := r.cli.Get(context.TODO(), request.NamespacedName, pod)
	if err != nil {
		return reconcile.Result{}, err
	}
	log.WithName("mock pod controller").Info("got pod", "podName", pod.Name, "podPhase", pod.Status.Phase)
	pod = pod.DeepCopy()
	if pod.Status.Phase == v1.PodPending {
		pod.Status.Phase = v1.PodRunning
		r.getLogger().Info("starting pod", "podName", pod.Name)
		err := r.cli.Status().Update(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	if pod.Status.Phase == v1.PodRunning {
		pod.Status.Phase = v1.PodSucceeded
		r.getLogger().Info("pod succeeded", "podName", pod.Name)
		err := r.cli.Status().Update(context.TODO(), pod)
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

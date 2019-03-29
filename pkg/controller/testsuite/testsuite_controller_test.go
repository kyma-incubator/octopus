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
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sync"
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

const defaultAssertionTimeout = time.Second * 20

func TestReconcileClusterTestSuite(t *testing.T) {

	t.Run("repeat, no concurrency", func(t *testing.T) {
		// GIVEN
		// Setup the Manager and Controller
		mgr, err := manager.New(cfg, manager.Options{})
		require.NoError(t, err)
		c := mgr.GetClient()

		testNs := generateTestNs()
		ctx := context.Background()

		suite := &testingv1alpha1.ClusterTestSuite{
			ObjectMeta: metav1.ObjectMeta{Name: "suite-test-repeat"},
			Spec: testingv1alpha1.TestSuiteSpec{
				Concurrency: 1,
				Count:       2,
			},
		}

		err = c.Create(ctx, suite)
		require.NoError(t, err)

		defer cleanupK8sObject(ctx, c, suite)

		require.NoError(t, add(mgr, newReconciler(mgr)))
		stopMgr, mgrStopped := StartTestManager(t, mgr)

		defer func() {
			close(stopMgr)
			mgrStopped.Wait()
		}()

		logf.SetLogger(logf.ZapLogger(false))

		podReconciler, err := startMockPodController(mgr, 0)
		require.NoError(t, err)

		// WHEN
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNs,
			},
		}
		err = c.Create(ctx, ns)
		require.NoError(t, err)

		defer cleanupK8sObject(ctx, c, ns)
		err = c.Create(ctx, getConcurrentTest("test-a", testNs))
		require.NoError(t, err)

		defer cleanupK8sObject(ctx, c, getConcurrentTest("test-a", testNs))

		err = c.Create(ctx, getSequentialTest("test-b", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getSequentialTest("test-b", testNs))
		// THEN
		repeat.FuncAtMost(t, func() error {
			return checkIfsuiteIsSucceeded(ctx, c, "suite-test-repeat")
		}, defaultAssertionTimeout)

		repeat.FuncAtMost(t, func() error {
			return checkIfPodsWereCreated(ctx, c, testNs, []string{
				"oct-tp-suite-test-repeat-test-a-0",
				"oct-tp-suite-test-repeat-test-a-1",
				"oct-tp-suite-test-repeat-test-b-0",
				"oct-tp-suite-test-repeat-test-b-1"})
		}, defaultAssertionTimeout)

		assertThatPodsCreatedSequentially(t, podReconciler.getAppliedChanges())
	})

	t.Run("repeat and concurrency", func(t *testing.T) {
		// GIVEN

		// Setup the Manager and Controller
		mgr, err := manager.New(cfg, manager.Options{})
		require.NoError(t, err)
		c := mgr.GetClient()

		testNs := generateTestNs()
		ctx := context.Background()

		suite := &testingv1alpha1.ClusterTestSuite{
			ObjectMeta: metav1.ObjectMeta{Name: "suite-test-concurrency"},
			Spec: testingv1alpha1.TestSuiteSpec{
				Concurrency: 3,
				Count:       3,
			},
		}
		err = c.Create(ctx, suite)
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, suite)

		require.NoError(t, add(mgr, newReconciler(mgr)))
		stopMgr, mgrStopped := StartTestManager(t, mgr)

		defer func() {
			close(stopMgr)
			mgrStopped.Wait()
		}()

		logf.SetLogger(logf.ZapLogger(false))

		podReconciler, err := startMockPodController(mgr, time.Millisecond*200)
		require.NoError(t, err)

		// WHEN
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNs,
			},
		}
		err = c.Create(ctx, ns)
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, ns)

		err = c.Create(ctx, getConcurrentTest("test-conc-a", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getConcurrentTest("test-conc-a", testNs))

		err = c.Create(ctx, getConcurrentTest("test-conc-b", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getConcurrentTest("test-conc-b", testNs))

		// THEN
		repeat.FuncAtMost(t, func() error {
			return checkIfsuiteIsSucceeded(ctx, c, "suite-test-concurrency")
		}, defaultAssertionTimeout)

		repeat.FuncAtMost(t, func() error {
			return checkIfPodsWereCreated(ctx, c, testNs, []string{
				"oct-tp-suite-test-concurrency-test-conc-a-0",
				"oct-tp-suite-test-concurrency-test-conc-a-1",
				"oct-tp-suite-test-concurrency-test-conc-a-2",
				"oct-tp-suite-test-concurrency-test-conc-b-0",
				"oct-tp-suite-test-concurrency-test-conc-b-1",
				"oct-tp-suite-test-concurrency-test-conc-b-2"})
		}, defaultAssertionTimeout)

		require.Len(t, podReconciler.getAppliedChanges(), 12)
		assertThatPodsCreatedConcurrently(t, podReconciler.getAppliedChanges())
	})

	t.Run("high concurrency but no concurrent tests", func(t *testing.T) {
		// GIVEN
		// Setup the Manager and Controller
		mgr, err := manager.New(cfg, manager.Options{})
		require.NoError(t, err)
		c := mgr.GetClient()

		testNs := generateTestNs()
		ctx := context.Background()

		suite := &testingv1alpha1.ClusterTestSuite{
			ObjectMeta: metav1.ObjectMeta{Name: "suite-test-sequential"},
			Spec: testingv1alpha1.TestSuiteSpec{
				Concurrency: 10,
				Count:       1,
			},
		}
		err = c.Create(ctx, suite)
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, suite)

		require.NoError(t, add(mgr, newReconciler(mgr)))
		stopMgr, mgrStopped := StartTestManager(t, mgr)
		defer func() {
			close(stopMgr)
			mgrStopped.Wait()
		}()

		logf.SetLogger(logf.ZapLogger(false))

		podReconciler, err := startMockPodController(mgr, 0)
		require.NoError(t, err)

		// WHEN
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNs,
			},
		}
		err = c.Create(ctx, ns)
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, ns)

		err = c.Create(ctx, getSequentialTest("test-a", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getSequentialTest("test-a", testNs))

		err = c.Create(ctx, getSequentialTest("test-b", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getSequentialTest("test-b", testNs))

		err = c.Create(ctx, getSequentialTest("test-c", testNs))
		require.NoError(t, err)
		defer cleanupK8sObject(ctx, c, getSequentialTest("test-c", testNs))

		// THEN
		repeat.FuncAtMost(t, func() error {
			return checkIfsuiteIsSucceeded(ctx, c, "suite-test-sequential")
		}, defaultAssertionTimeout)

		repeat.FuncAtMost(t, func() error {
			return checkIfPodsWereCreated(ctx, c, testNs, []string{
				"oct-tp-suite-test-sequential-test-a-0",
				"oct-tp-suite-test-sequential-test-b-0",
				"oct-tp-suite-test-sequential-test-c-0"})
		}, defaultAssertionTimeout)

		assertThatPodsCreatedSequentially(t, podReconciler.getAppliedChanges())
	})


}

func assertThatPodsCreatedConcurrently(t *testing.T, appliedChanges []podStatusChanges) {
	require.True(t, len(appliedChanges)%2 == 0, "expected even number of applied pod changes [%d]", len(appliedChanges))
	changesOrder := make(map[string][]int, 0)
	for idx, ch := range appliedChanges {
		changesOrder[ch.podName] = append(changesOrder[ch.podName], idx)
	}

	interchangedPodChanges := false
	for _, v := range changesOrder {
		require.Len(t, v, 2)
		if v[1]-v[0] != 1 {
			// at least 2 pods were running concurrently
			interchangedPodChanges = true
			break
		}
	}

	assert.True(t, interchangedPodChanges)

}

func assertThatPodsCreatedSequentially(t *testing.T, appliedChanges []podStatusChanges) {
	require.True(t, len(appliedChanges)%2 == 0, "expected even number of applied pod changes [%d]", len(appliedChanges))
	for i := 0; i < len(appliedChanges); i += 2 {
		assert.True(t, appliedChanges[i].podName == appliedChanges[i+1].podName)
		assert.Equal(t, v1.PodRunning, appliedChanges[i].phase)
		assert.Equal(t, v1.PodSucceeded, appliedChanges[i+1].phase)

	}
}

func checkIfsuiteIsSucceeded(ctx context.Context, reader client.Reader, suiteName string) error {
	var actualSuite testingv1alpha1.ClusterTestSuite
	if err := reader.Get(ctx, types.NamespacedName{Name: suiteName}, &actualSuite); err != nil {
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
}

func checkIfPodsWereCreated(ctx context.Context, reader client.Reader, ns string, expectedPodNames []string) error {
	actualPods := &v1.PodList{}
	if err := reader.List(ctx, &client.ListOptions{
		Namespace: ns,
	}, actualPods); err != nil {
		return err
	}

	if len(actualPods.Items) != len(expectedPodNames) {
		return fmt.Errorf("wrong number of pods, expected 4, was [%d]", len(actualPods.Items))
	}

	var actualPodNames = make(map[string]bool, len(actualPods.Items))
	for _, p := range actualPods.Items {
		actualPodNames[p.Name] = true
	}

	for _, exp := range expectedPodNames {
		if !actualPodNames[exp] {
			return fmt.Errorf("missing pod [%s], actualPodNames pods: [%v]", exp, actualPodNames)
		}
	}
	return nil
}

func cleanupK8sObject(ctx context.Context, writer client.Writer, object runtime.Object) {
	writer.Delete(ctx, object, client.GracePeriodSeconds(0))
}

// We need to manually change status of Pod
type mockPodReconciler struct {
	cli client.Client
	// mutex protecting access to applied changes slice
	mtx sync.Mutex
	// chronologically list of pod changes. Use getter to access it autside the struct
	appliedChanges []podStatusChanges
	finishPodAfter time.Duration
}

func (r *mockPodReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
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
		if err := r.cli.Status().Update(ctx, pod); err != nil {
			return reconcile.Result{}, err
		}
		r.appliedChanges = append(r.appliedChanges, podStatusChanges{podName: pod.Name, phase: pod.Status.Phase, ts: time.Now()})
		return reconcile.Result{RequeueAfter: r.finishPodAfter}, nil
	}
	if pod.Status.Phase == v1.PodRunning {
		var runningSince time.Time
		for _, ch := range r.appliedChanges {
			if ch.podName == pod.Name && ch.phase == v1.PodRunning {
				runningSince = ch.ts
			}
		}
		now := time.Now()
		whenShouldFinishPod := runningSince.Add(r.finishPodAfter)

		if now.Before(whenShouldFinishPod) {
			return reconcile.Result{RequeueAfter: whenShouldFinishPod.Sub(now)}, nil
		}
		pod.Status.Phase = v1.PodSucceeded
		r.getLogger().Info("pod succeeded", "podName", pod.Name)
		err := r.cli.Status().Update(ctx, pod)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.appliedChanges = append(r.appliedChanges, podStatusChanges{podName: pod.Name, phase: pod.Status.Phase, ts: time.Now()})
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *mockPodReconciler) getAppliedChanges() []podStatusChanges {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.appliedChanges
}

func (r *mockPodReconciler) getLogger() logr.Logger {
	return logf.Log.WithName("mock pod controller")
}

func startMockPodController(mgr manager.Manager, finishPodAfter time.Duration) (*mockPodReconciler, error) {
	pr := &mockPodReconciler{
		cli:            mgr.GetClient(),
		mtx:            sync.Mutex{},
		finishPodAfter: finishPodAfter,
	}
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: pr})
	if err != nil {
		return nil, err
	}

	err = c.Watch(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		GenericFunc: func(ev event.GenericEvent) bool {
			return ev.Meta.GetLabels()[testingv1alpha1.LabelKeyCreatedByOctopus] == "true"
		},
	})
	if err != nil {
		return nil, err
	}

	return pr, nil
}

type podStatusChanges struct {
	podName string
	phase   v1.PodPhase
	ts      time.Time
}

func getConcurrentTest(testName, ns string) *testingv1alpha1.TestDefinition {
	return &testingv1alpha1.TestDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: ns},
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

func getSequentialTest(testName, ns string) *testingv1alpha1.TestDefinition {
	return &testingv1alpha1.TestDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: testName, Namespace: ns},
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

func generateTestNs() string {
	suffix := rand.String(10)
	return fmt.Sprintf("testing-octopus-%s", suffix)
}

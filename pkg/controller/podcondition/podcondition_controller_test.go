/*
Copyright 2019 Itay Shakury.

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

package podcondition

import (
	"testing"
	"time"

	conditionerv1alpha1 "github.com/itaysk/kube-conditioner/pkg/apis/conditioner/v1alpha1"
	"github.com/itaysk/kube-conditioner/pkg/datasource/prometheus"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "testcondition", Namespace: "default"}}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &conditionerv1alpha1.PodCondition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testcondition",
			Namespace: "default",
		},
		Spec: conditionerv1alpha1.PodConditionSpec{
			Interval: 1000, //TODO: this should be optional, use deafult
			LabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"testing": "123",
				},
			},
			PrometheusSource: &conditionerv1alpha1.PrometheusSource{
				ServerURL: "http://localhost:9090", //TODO: mock prometheus or add real to test setup
				Rule:      "vector(1) > 0",
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	r := newReconciler(mgr).(*ReconcilePodCondition)
	recFn, requests := SetupTestReconcile(r)
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer func() {
		c.Delete(context.TODO(), instance)
	}()

	//this tells us the reconciler was indeed triggered by the created instance
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	wKey := types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}
	g.Expect(r.workers).To(gomega.HaveKey(wKey))
	g.Expect(r.workers[wKey].conditionName).To(gomega.Equal(instance.Name))
	g.Expect(r.workers[wKey].ds.(*prometheus.PrometheusSource).ServerURL).To(gomega.Equal(instance.Spec.PrometheusSource.ServerURL))
	g.Expect(r.workers[wKey].ds.(*prometheus.PrometheusSource).Rule).To(gomega.Equal(instance.Spec.PrometheusSource.Rule))
	g.Expect(r.workers[wKey].labelSelector.MatchLabels).To(gomega.Equal(instance.Spec.LabelSelector.MatchLabels)) //TODO: remove assumption on label usage here
	//TODO: how to inspect ticker interval?

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	// g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	// g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	// g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
	// 	Should(gomega.Succeed())

	// Manually delete Deployment since GC isn't enabled in the test control plane
	// g.Eventually(func() error { return c.Delete(context.TODO(), deploy) }, timeout).
	// 	Should(gomega.MatchError("deployments.apps \"foo-deployment\" not found"))

}

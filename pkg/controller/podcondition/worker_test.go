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

	"github.com/itaysk/kube-conditioner/pkg/datasource"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var podTemplate = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "testpod",
		Namespace: "default",
		Labels: map[string]string{
			"testing": "123",
		},
	},
	Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Name:  "testpod",
				Image: "alpine",
			},
		},
	},
}

func TestUpdateCondition(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Fatal("cant initialize client", err)
	}

	//create 3 pods, 2 of them matching the label selector
	pod1 := podTemplate.DeepCopy()
	pod1.ObjectMeta.Name = "testpod1"
	pod2 := podTemplate.DeepCopy()
	pod2.ObjectMeta.Name = "testpod2"
	pod3 := podTemplate.DeepCopy()
	pod3.ObjectMeta.Name = "testpod3"
	delete(pod3.ObjectMeta.Labels, "testing")
	//there's a seperate test for storage, so maybe can ignore errors here
	c.Create(context.TODO(), pod1)
	c.Create(context.TODO(), pod2)
	c.Create(context.TODO(), pod3)
	defer func() {
		c.Delete(context.TODO(), pod1)
		c.Delete(context.TODO(), pod2)
		c.Delete(context.TODO(), pod3)
	}()

	conditionName := "testcondition"
	w, err := NewWorker(
		conditionName,
		metav1.LabelSelector{MatchLabels: map[string]string{"testing": "123"}},
		time.Duration(1),
		&datasource.DataSourceMock{Result: true}, //TODO: no need for prometheus in the instance here? also, updateconditino is being passed true
		c)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	w.updateCondition(true)

	pod1Res := corev1.Pod{}
	pod2Res := corev1.Pod{}
	pod3Res := corev1.Pod{}
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod1"}, &pod1Res)
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod2"}, &pod2Res)
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod3"}, &pod3Res)

	g.Expect(pod1Res.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod1Res.Status.Conditions[0].Type).To(gomega.Equal(corev1.PodConditionType(conditionName)))
	g.Expect(pod1Res.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionStatus("true")))
	g.Expect(pod1Res.Status.Conditions[0].LastProbeTime).To(gomega.Equal(pod1Res.Status.Conditions[0].LastTransitionTime))

	g.Expect(pod2Res.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod2Res.Status.Conditions[0].Type).To(gomega.Equal(corev1.PodConditionType(conditionName)))
	g.Expect(pod2Res.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionStatus("true")))
	g.Expect(pod2Res.Status.Conditions[0].LastProbeTime).To(gomega.Equal(pod2Res.Status.Conditions[0].LastTransitionTime))

	g.Expect(pod3Res.Status.Conditions).To(gomega.HaveLen(0))

	// this part may be replaced by a test of the ticker as part of a test for the Start function
	time.Sleep(time.Duration(time.Second * 1)) //allow some time difference between lastprobe and lasttransition

	w.updateCondition(true)

	pod1Res = corev1.Pod{}
	pod2Res = corev1.Pod{}
	pod3Res = corev1.Pod{}
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod1"}, &pod1Res)
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod2"}, &pod2Res)
	c.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "testpod3"}, &pod3Res)

	g.Expect(pod1Res.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod1Res.Status.Conditions[0].Type).To(gomega.Equal(corev1.PodConditionType(conditionName)))
	g.Expect(pod1Res.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionStatus("true")))
	g.Expect(pod1Res.Status.Conditions[0].LastProbeTime).ToNot(gomega.Equal(pod1Res.Status.Conditions[0].LastTransitionTime))

	g.Expect(pod2Res.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod2Res.Status.Conditions[0].Type).To(gomega.Equal(corev1.PodConditionType(conditionName)))
	g.Expect(pod2Res.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionStatus("true")))
	g.Expect(pod2Res.Status.Conditions[0].LastProbeTime).ToNot(gomega.Equal(pod2Res.Status.Conditions[0].LastTransitionTime))

	g.Expect(pod3Res.Status.Conditions).To(gomega.HaveLen(0))
}

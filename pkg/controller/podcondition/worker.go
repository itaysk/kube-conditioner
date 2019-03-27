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
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/itaysk/kube-conditioner/pkg/datasource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Worker is a single control loop in charge if a condition
type Worker struct {
	conditionName string
	labelSelector metav1.LabelSelector
	ticker        *time.Ticker
	cancel        chan struct{}
	kubeClient    client.Client
	ds            datasource.DataSource
}

//NewWorker creates a new worker
func NewWorker(name string, labelSelector metav1.LabelSelector, interval time.Duration, ds datasource.DataSource, kubeClient client.Client) (*Worker, error) {
	//TODO: return the cancel channel

	w := &Worker{
		conditionName: name,
		labelSelector: labelSelector,
		ticker:        time.NewTicker(interval),
		cancel:        make(chan struct{}),
		kubeClient:    kubeClient,
	}
	err := w.SetDataSource(ds)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// SetDataSource sets the datasource member if it's valid, or kills the worker if its not
func (w *Worker) SetDataSource(ds datasource.DataSource) error {
	if ds == nil {
		w.cancel <- struct{}{}
		return errors.New("datasource invalid")
	}
	w.ds = ds
	return nil
}

// Start starts the worker loop
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.ticker.C:
				val, err := w.ds.Evaluate()
				if err != nil {
					w.cancel <- struct{}{} //TODO: retry + circuit braker
					//TODO: if not canceled by ^^, evaluate will be called continuously instead of at interval, why?
				} else {
					w.updateCondition(val)
				}
			case <-w.cancel:
				w.ticker.Stop()
			}
		}
	}()
}

func (w *Worker) updateCondition(conditionValue bool) {
	//TODO: what is multiple workers try to update same condition on same pod? for now just let ot fail

	pods := &corev1.PodList{}
	labelSelector, _ := metav1.LabelSelectorAsSelector(&w.labelSelector)
	err := w.kubeClient.List(context.TODO(), &client.ListOptions{LabelSelector: labelSelector}, pods)
	if err != nil {
		log.Error(err, "error listing pods")
	}

	conditionType := corev1.PodConditionType(w.conditionName)
	conditionStatus := corev1.ConditionStatus(strconv.FormatBool(conditionValue))
	for _, pod := range pods.Items {
		now := metav1.NewTime(time.Now())
		found := false
		for i, condition := range pod.Status.Conditions {
			if condition.Type == conditionType {
				found = true
				condition.LastProbeTime = now
				if condition.Status != conditionStatus {
					condition.LastTransitionTime = now
					condition.Status = conditionStatus
				}
				pod.Status.Conditions[i] = condition
			}
		}
		if !found {
			pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
				Type:               conditionType,
				Status:             conditionStatus,
				LastProbeTime:      now,
				LastTransitionTime: now,
			})
		}

		//update regardless of existing condition just to refresh the lastprobetime
		err = w.kubeClient.Status().Update(context.TODO(), &pod)
		if err != nil {
			log.Error(err, "error updating condition")
		}
	}

}

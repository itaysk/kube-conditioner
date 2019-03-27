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
	"time"

	conditionerv1alpha1 "github.com/itaysk/kube-conditioner/pkg/apis/conditioner/v1alpha1"
	"github.com/itaysk/kube-conditioner/pkg/datasource"
	"github.com/itaysk/kube-conditioner/pkg/datasource/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

// Add creates a new Conditioner Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodCondition{Client: mgr.GetClient(), scheme: mgr.GetScheme(), workers: make(map[apitypes.NamespacedName]*Worker)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podcondition-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Conditioner
	err = c.Watch(&source.Kind{Type: &conditionerv1alpha1.PodCondition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePodCondition{}

// ReconcilePodCondition reconciles a Conditioner object
type ReconcilePodCondition struct {
	client.Client
	scheme  *runtime.Scheme
	workers map[apitypes.NamespacedName]*Worker
}

// Reconcile is the handler of kubernetes watch events for the controller. It acts based on the recieved request.
// TODO: remove unneeded permissions
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=conditioner.itaysk.com,resources=podconditions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=conditioner.itaysk.com,resources=podconditions/status,verbs=get;update;patch
func (r *ReconcilePodCondition) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &conditionerv1alpha1.PodCondition{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	ds := resolveDatasource(instance.Spec) //TODO: better to remove the worker if no datasource

	//TODO: handle delete
	worker, ok := r.workers[request.NamespacedName]
	if !ok {
		w, err := NewWorker(
			instance.Name,
			instance.Spec.LabelSelector,
			time.Duration(time.Duration(instance.Spec.Interval)*time.Millisecond),
			ds,
			r.Client,
		)
		if err != nil {
			log.Error(err, "error creating worker")
			//w and worker should be nil in this case
		} else {
			r.workers[request.NamespacedName] = w
			w.Start()
		}
	} else {
		worker.labelSelector = *instance.Spec.LabelSelector.DeepCopy()
		err := worker.SetDataSource(ds)
		if err != nil {
			log.Error(err, "error setting datasource")
			// worker shoud be a struct in this case
			worker.cancel <- struct{}{}
		}
		//TODO: support changing interval by recreating the worker(ticker)
	}

	return reconcile.Result{}, nil
}

//TODO: datasource management needs refactoring
func resolveDatasource(spec conditionerv1alpha1.PodConditionSpec) datasource.DataSource {
	if spec.PrometheusSource != nil {
		return prometheus.NewPrometheusSource(spec.PrometheusSource.ServerURL, spec.PrometheusSource.Rule)
	}
	return nil
}

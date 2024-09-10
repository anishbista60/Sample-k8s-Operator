/*
Copyright 2024 Anish Bista.

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

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1alpha1 "github.com/anishbista60/k8s-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.operator.anishbista.xyz,resources=scalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.operator.anishbista.xyz,resources=scalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.operator.anishbista.xyz,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	scaler := &demov1alpha1.Scaler{}
	if err := r.Get(ctx, req.NamespacedName, scaler); err != nil {
		return ctrl.Result{}, nil
	}

	start := scaler.Spec.Start
	end := scaler.Spec.End
	replicas := scaler.Spec.Replicas

	currenthours := time.Now().UTC().Hour()
	fmt.Printf("The current hours is %v\n", currenthours)
	if currenthours >= start && currenthours <= end {
		if err := scaledeployment(scaler, ctx, r, replicas); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func scaledeployment(scaler *demov1alpha1.Scaler, ctx context.Context, r *ScalerReconciler, replicas int32) error {
	for _, deploy := range scaler.Spec.Deployments {
		deployment := &appsv1.Deployment{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: deploy.Namespace,
			Name:      deploy.Name,
		}, deployment); err != nil {
			return err
		}
		if deployment.Spec.Replicas != &replicas {
			deployment.Spec.Replicas = &replicas
			if err := r.Update(ctx, deployment); err != nil {
				return err
			}
			if err := r.Status().Update(ctx, scaler); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.Scaler{}).
		Complete(r)
}

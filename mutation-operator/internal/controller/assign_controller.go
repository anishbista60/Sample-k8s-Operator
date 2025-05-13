/*
Copyright 2024.

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
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mutationsv1 "github.com/kratos/kratos-operator/api/v1alpha1"
)

// AssignReconciler reconciles an Assign object
type AssignReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mutations.kratos.dev,resources=assigns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mutations.kratos.dev,resources=assigns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mutations.kratos.dev,resources=assigns/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;update;patch;create
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

// Reconcile reconciles the Assign resource
func (r *AssignReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Assign resource", "namespace", req.Namespace, "name", req.Name)

	// Fetch the Assign instance
	var assign mutationsv1.Assign
	if err := r.Get(ctx, req.NamespacedName, &assign); err != nil {
		log.Error(err, "Unable to fetch Assign resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Apply mutations to existing Pods
	if err := r.applyMutation(ctx, assign); err != nil {
		log.Error(err, "Failed to apply mutation")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled Assign resource")
	return ctrl.Result{RequeueAfter: 6 * time.Second}, nil
}

// applyMutation applies the mutation defined in the Assign resource
func (r *AssignReconciler) applyMutation(ctx context.Context, assign mutationsv1.Assign) error {
	log := log.FromContext(ctx)

	// List all Pods in the target namespace
	var podList corev1.PodList
	namespace := assign.Spec.Match.NameSpaceSelector.MatchLabels["kubernetes.io/metadata.name"]

	// Validate namespace selection
	if namespace == "" {
		return fmt.Errorf("namespace selector is not defined in Assign resource")
	}

	// Fetch pods in the specified namespace
	if err := r.List(ctx, &podList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range podList.Items {
		// Fetch the ReplicaSet
		replicaSet := &appsv1.ReplicaSet{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: pod.Namespace,
			Name:      pod.OwnerReferences[0].Name,
		}, replicaSet); err != nil {
			log.Error(err, "Failed to get ReplicaSet for Pod", "pod", pod.Name)
			continue
		}

		// Fetch the Deployment from the ReplicaSet
		deploymentName := replicaSet.OwnerReferences[0].Name
		deployment := &appsv1.Deployment{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: pod.Namespace,
			Name:      deploymentName,
		}, deployment); err != nil {
			log.Error(err, "Failed to get Deployment for ReplicaSet", "replicaSet", replicaSet.Name)
			continue
		}

		// Apply mutations
		if err := r.mutateDeployment(ctx, deployment, assign); err != nil {
			log.Error(err, "Failed to mutate deployment", "deployment", deployment.Name)
			return err
		}
	}
	return nil
}

// mutateDeployment applies mutations to the Deployment with retry logic
func (r *AssignReconciler) mutateDeployment(ctx context.Context, deployment *appsv1.Deployment, assign mutationsv1.Assign) error {
	log := log.FromContext(ctx)
	updated := false

	// Retry logic for conflicts
	return wait.ExponentialBackoff(wait.Backoff{
		Steps:    5,
		Duration: 200 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}, func() (bool, error) {
		// Fetch the latest state
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		}, deployment); err != nil {
			return false, err
		}

		// Apply nodeSelector mutation
		if assign.Spec.Location == "spec.nodeSelector" {
			var nodeSelector map[string]string
			if err := json.Unmarshal(assign.Spec.Parameters.Assign.Value.Raw, &nodeSelector); err != nil {
				return false, fmt.Errorf("failed to unmarshal nodeSelector value: %w", err)
			}

			if deployment.Spec.Template.Spec.NodeSelector == nil {
				deployment.Spec.Template.Spec.NodeSelector = make(map[string]string)
			}

			for k, v := range nodeSelector {
				if deployment.Spec.Template.Spec.NodeSelector[k] != v {
					deployment.Spec.Template.Spec.NodeSelector[k] = v
					updated = true
				}
			}
		}

		// Apply tolerations mutation
		if assign.Spec.Location == "spec.tolerations" {
			var tolerations []corev1.Toleration
			if err := json.Unmarshal(assign.Spec.Parameters.Assign.Value.Raw, &tolerations); err != nil {
				return false, fmt.Errorf("failed to unmarshal tolerations value: %w", err)
			}

			deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, tolerations...)
			updated = true
		}

		// Patch Deployment if changes were made
		if updated {
			if err := r.Update(ctx, deployment); err != nil {
				if errors.IsConflict(err) {
					log.Info("Conflict detected, retrying update", "deployment", deployment.Name)
					return false, nil
				}
				return false, err
			}
			log.Info("Successfully updated Deployment", "deployment", deployment.Name)
		}
		return true, nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *AssignReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mutationsv1.Assign{}).
		Named("assign").
		Complete(r)
}

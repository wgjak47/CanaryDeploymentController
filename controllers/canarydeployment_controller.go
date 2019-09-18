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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webappv1 "CanaryDeploymentController/api/v1"
)

var (
	deploymentOwnerKey = ".metadata.controller"
)

// CanaryDeploymentReconciler reconciles a CanaryDeployment object
type CanaryDeploymentReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=webapp.wgjak47.me,resources=canarydeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.wgjak47.me,resources=canarydeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
func (r *CanaryDeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("canarydeployment", req.NamespacedName)

	// your logic here
	log.Info("fetching CanaryDeployment")
	canaryDeployment := webappv1.CanaryDeployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, &canaryDeployment); err != nil {
		log.Error(err, "failed to get canaryDeployment resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.removeUselessDeployments(ctx, log, &canaryDeployment); err != nil {
		log.Error(err, "removeOldDeployment error")
		return ctrl.Result{}, err
	}

	if newDeployments, err := r.generateNewDeployments(&canaryDeployment.Spec); err != nil {
		log.Error(err, "error generate new deployments")
		return ctrl.Result{}, err
	} else {
		for _, newDeployment := range newDeployments {
			oldDeployment := apps.Deployment{}
			err := r.Client.Get(ctx, client.ObjectKey{Namespace: newDeployment.Namespace, Name: newDeployment.Name}, &oldDeployment)
			if err != nil {
				if apierrors.IsNotFound(err) {
					log.Info("not found deployment, start creating", "name", newDeployment.Name)
					if err := r.Client.Create(ctx, &newDeployment); err != nil {
						log.Error(err, "failed to create Deployment resource", "name", newDeployment.Name)
						return ctrl.Result{}, err
					}
				}
				log.Error(err, "failed to get old Deployment", "name", newDeployment.Name)
				return ctrl.Result{}, err
			} else {
				if err := r.Client.Update(ctx, &newDeployment); err != nil {
					log.Error(err, "failed to update Deployment update", "name", newDeployment.Name)
					return ctrl.Result{}, err
				}
			}
			canaryDeployment.Status.AppStatus[newDeployment.Name] = &newDeployment.Status
		}
		if err := r.Client.Status().Update(ctx, &canaryDeployment); err != nil {
			log.Error(err, "failed to update MyKind status")
			return ctrl.Result{}, err
		}
		log.Info("resource status synced", "name", canaryDeployment.Name)
	}

	return ctrl.Result{}, nil
}

// Combine CommonSpec and AppSpecs to deployments
func (r *CanaryDeploymentReconciler) generateNewDeployments(spec *webappv1.CanaryDeploymentSpec) ([]apps.Deployment, error) {
	return []apps.Deployment{}, nil
}

// remove the useless deployment that not control by CanaryDeployment any more
func (r *CanaryDeploymentReconciler) removeUselessDeployments(
	ctx context.Context, log logr.Logger, cd *webappv1.CanaryDeployment) error {
	deployments := apps.DeploymentList{}
	if err := r.Client.List(
		ctx, &deployments,
		client.InNamespace(cd.Namespace),
		client.MatchingFields{deploymentOwnerKey: cd.Name},
	); err != nil {
		log.Error(err, "List error")
		return err
	}

	for _, deployment := range deployments.Items {
		if _, ok := cd.Spec.AppSpecs[deployment.Name]; !ok {
			if err := r.Client.Delete(ctx, &deployment); err != nil {
				log.Error(err, "delete Deployment failed", "name", deployment.Name)
				return err
			}
			log.Info("deleted Deployment", "name", deployment.Name)
		}
	}
	return nil
}

func (r *CanaryDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CanaryDeployment{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}

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
	webappv1 "CanaryDeploymentController/api/v1"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	if newDeployments, err := r.generateNewDeployments(&canaryDeployment.Spec, &canaryDeployment); err != nil {
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
			canaryDeployment.Status.AppStatus[newDeployment.Name] = newDeployment.Status
		}
		if err := r.Client.Status().Update(ctx, &canaryDeployment); err != nil {
			log.Error(err, "failed to update MyKind status")
			return ctrl.Result{}, err
		}
		log.Info("resource status synced", "name", canaryDeployment.Name)
	}

	return ctrl.Result{}, nil
}

func (r *CanaryDeploymentReconciler) patchDeploymentSpec(
	spec *apps.DeploymentSpec,
	patch *apps.DeploymentSpec,
	patchType string,
) (apps.DeploymentSpec, error) {

	originSpec, err := json.Marshal(spec)
	if err != nil {
		return apps.DeploymentSpec{}, err
	}

	patchSpec, err := json.Marshal(patch)
	if err != nil {
		return apps.DeploymentSpec{}, err
	}
	var NewDeploymentSpecBytes []byte
	if patchType == "strategic" {
		NewDeploymentSpecBytes, err = strategicpatch.StrategicMergePatch(originSpec, patchSpec, apps.DeploymentSpec{})
		if err != nil {
			return apps.DeploymentSpec{}, err
		}
	} else if patchType == "merge" {
		NewDeploymentSpecBytes, err = jsonpatch.MergePatch(originSpec, patchSpec)
		if err != nil {
			return apps.DeploymentSpec{}, err
		}
	} else {
		return apps.DeploymentSpec{}, errors.New("only support Merge or Strategic patch type")
	}
	newDeploymentSpec := apps.DeploymentSpec{}
	err = json.Unmarshal(NewDeploymentSpecBytes, &newDeploymentSpec)
	if err != nil {
		return apps.DeploymentSpec{}, err
	}
	return newDeploymentSpec, nil
}

func (r *CanaryDeploymentReconciler) buildDeployment(spec *apps.DeploymentSpec, cd *webappv1.CanaryDeployment, name string) apps.Deployment {
	newDeployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", cd.Name, name),
			Namespace:       cd.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cd, webappv1.GroupVersion.WithKind("CanaryDeployment"))},
		},
		Spec: *spec,
	}
	return newDeployment
}

// Combine CommonSpec and AppSpecs to deployments
func (r *CanaryDeploymentReconciler) generateNewDeployments(spec *webappv1.CanaryDeploymentSpec, cd *webappv1.CanaryDeployment) ([]apps.Deployment, error) {
	newDeployments := []apps.Deployment{}
	for name, patch := range spec.AppSpecs {
		newDeploymentSpec, err := r.patchDeploymentSpec(&spec.CommonSpec, &patch.Spec, patch.Type)
		if err != nil {
			return []apps.Deployment{}, err
		}
		newDeployment := r.buildDeployment(&newDeploymentSpec, cd, name)
		newDeployments = append(newDeployments, newDeployment)
	}
	return newDeployments, nil
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
	if err := mgr.GetFieldIndexer().IndexField(&apps.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the Deployment object, extract the owner...
		depl := rawObj.(*apps.Deployment)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a MyKind...
		if owner.APIVersion != webappv1.GroupVersion.String() || owner.Kind != "MyKind" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CanaryDeployment{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}

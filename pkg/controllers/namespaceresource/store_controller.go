/*
Copyright The Ratify Authors.
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

package namespaceresource

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	"github.com/ratify-project/ratify/internal/constants"
	"github.com/ratify-project/ratify/pkg/controllers"
	"github.com/ratify-project/ratify/pkg/controllers/utils"
	"github.com/sirupsen/logrus"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedstores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedstores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedstores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	storeLogger := logrus.WithContext(ctx)

	var store configv1beta1.NamespacedStore
	var resource = req.Name
	storeLogger.Infof("reconciling namspaced store '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {
		if apierrors.IsNotFound(err) {
			storeLogger.Infof("deletion detected, removing store %v", req.Name)
			controllers.NamespacedStores.DeleteStore(req.Namespace, resource)
		} else {
			storeLogger.Error(err, "unable to fetch store")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := storeAddOrReplace(store.Spec, resource, req.Namespace); err != nil {
		storeLogger.Error(err, "unable to create store from store crd")
		writeStoreStatus(ctx, r, &store, storeLogger, false, err.Error())
		return ctrl.Result{}, err
	}

	writeStoreStatus(ctx, r, &store, storeLogger, true, "")

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.NamespacedStore{}).
		Complete(r)
}

// Creates a store reference from CRD spec and add store to map
func storeAddOrReplace(spec configv1beta1.NamespacedStoreSpec, fullname, namespace string) error {
	storeConfig, err := utils.CreateStoreConfig(spec.Parameters.Raw, spec.Name, spec.Source)
	if err != nil {
		return fmt.Errorf("unable to convert store spec to store config, err: %w", err)
	}

	return utils.UpsertStoreMap(spec.Version, spec.Address, fullname, namespace, storeConfig)
}

func writeStoreStatus(ctx context.Context, r client.StatusClient, store *configv1beta1.NamespacedStore, logger *logrus.Entry, isSuccess bool, errorString string) {
	if isSuccess {
		store.Status.IsSuccess = true
		store.Status.Error = ""
		store.Status.BriefError = ""
	} else {
		store.Status.IsSuccess = false
		store.Status.Error = errorString
		if len(errorString) > constants.MaxBriefErrLength {
			store.Status.BriefError = fmt.Sprintf("%s...", errorString[:constants.MaxBriefErrLength])
		}
	}

	if statusErr := r.Status().Update(ctx, store); statusErr != nil {
		logger.Error(statusErr, ",unable to update store error status")
	}
}

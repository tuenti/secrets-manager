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
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	smv1alpha1 "github.com/tuenti/secrets-manager/api/v1alpha1"
	"github.com/tuenti/secrets-manager/backend"
)

const (
	// https://golang.org/pkg/time/#pkg-constants
	timestampFormat = "2006-01-02T15.04.05Z"
	finalizerName   = "secret.finalizer.secrets-manager.tuenti.io"
	allNamespaces   = "all"
)

// SecretDefinitionReconciler reconciles a SecretDefinition object
type SecretDefinitionReconciler struct {
	client.Client
	Backend              backend.Client
	Log                  logr.Logger
	Ctx                  context.Context
	APIReader            client.Reader
	ReconciliationPeriod time.Duration
	WatchNamespaces      map[string]bool
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// Ignore not found errors
func ignoreNotFoundError(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// isNotMarkedForRemoval will determine if the SecretDefinition object has been marked to be deleted
func isNotMarkedForRemoval(sDef smv1alpha1.SecretDefinition) bool {
	return sDef.ObjectMeta.DeletionTimestamp.IsZero()
}

// getDesiredState reads the content from the Datasource for later comparison
func (r *SecretDefinitionReconciler) getDesiredState(keysMap map[string]smv1alpha1.DataSource) (map[string][]byte, error) {
	desiredState := make(map[string][]byte)
	var err error
	for k, v := range keysMap {
		bSecret, err := r.Backend.ReadSecret(v.Path, v.Key)
		if err != nil {
			r.Log.Error(err, "unable to read secret from backend", "path", v.Path, "key", v.Key)
			return nil, err
		}
		decoder, err := backend.NewDecoder(v.Encoding)
		if err != nil {
			r.Log.Error(err, "refusing to use encoding", "encoding", v.Encoding)
			return nil, err
		}
		desiredState[k], err = decoder.DecodeString(bSecret)
		if err != nil {
			r.Log.Error(err, "unable to decode data for secret", "encoding", v.Encoding, "path", v.Path, "key", v.Key)
			return nil, err
		}
	}
	return desiredState, err
}

// getCurrentState reads the content from the Kubernetes Secret API object for later comparison
func (r *SecretDefinitionReconciler) getCurrentState(namespace string, name string) (map[string][]byte, error) {
	// We don't read secrets from cache, as it's not the object we reconcile
	reader := r.APIReader
	data := make(map[string][]byte)
	secret := &corev1.Secret{}
	err := reader.Get(r.Ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
	if err != nil {
		secretReadErrorsTotal.WithLabelValues(name, namespace).Inc()
		return data, err
	}
	data = secret.Data
	return data, err
}

// upsertSecret will create or update a secret
func (r *SecretDefinitionReconciler) upsertSecret(sDef *smv1alpha1.SecretDefinition, data map[string][]byte) error {
	secret := &corev1.Secret{
		Type: corev1.SecretType(sDef.Spec.Type),
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sDef.Namespace,
			Labels: map[string]string{
				"managedBy":     "secrets-manager",
				"lastUpdatedAt": time.Now().Format(timestampFormat),
			},
			Name: sDef.Spec.Name,
		},
		Data: data,
	}
	err := r.Create(r.Ctx, secret)
	if errors.IsAlreadyExists(err) {
		err = r.Update(r.Ctx, secret)
	}
	return err
}

// deleteSecret will delete a secret given its namespace and name
func (r *SecretDefinitionReconciler) deleteSecret(namespace string, name string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return r.Delete(r.Ctx, secret)
}

// shouldWatch will return true if the secretDefinition is in a watchable namespace
func (r *SecretDefinitionReconciler) shouldWatch(sDefNamespace string) bool {
	if len(r.WatchNamespaces) > 0 {
		return r.WatchNamespaces[sDefNamespace]
	}
	return true
}

// AddFinalizerIfNotPresent will check if finalizerName is the finalizers slice
func (r *SecretDefinitionReconciler) AddFinalizerIfNotPresent(sDef *smv1alpha1.SecretDefinition, finalizerName string) error {
	if !containsString(sDef.ObjectMeta.Finalizers, finalizerName) {
		sDef.ObjectMeta.Finalizers = append(sDef.ObjectMeta.Finalizers, finalizerName)
		return r.Update(r.Ctx, sDef)
	}
	return nil
}

// +kubebuilder:rbac:groups=secrets-manager.tuenti.io,resources=secretdefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secrets-manager.tuenti.io,resources=secretdefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *SecretDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secretdefinition", req.NamespacedName)

	sDef := &smv1alpha1.SecretDefinition{}

	err := r.Get(r.Ctx, req.NamespacedName, sDef)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not get SecretDefinition '%s'", req.NamespacedName))
		return ctrl.Result{}, ignoreNotFoundError(err)
	}

	secretName := sDef.Spec.Name
	secretNamespace := sDef.Namespace

	log = log.WithValues("secret", fmt.Sprintf("%s/%s", secretNamespace, secretName))

	if isNotMarkedForRemoval(*sDef) {

		err = r.AddFinalizerIfNotPresent(sDef, finalizerName)
		if err != nil {
			log.Error(err, "unable to update SecretDefinition finalizers", "finalizer", finalizerName)
			return ctrl.Result{}, err
		}

		if !r.shouldWatch(sDef.Namespace) {
			log.Info("outside watched namespaces, ignoring", "watched_namespaces", r.WatchNamespaces)
			return ctrl.Result{}, nil
		}
		// Get data from the secret source of truth
		desiredState, err := r.getDesiredState(sDef.Spec.KeysMap)

		if err != nil {
			log.Error(err, "unable to get desired state for secret")
			secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
			secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
			return ctrl.Result{}, err
		}

		// Get the actual secret from Kubernetes
		currentState, err := r.getCurrentState(secretNamespace, secretName)

		if err != nil && !errors.IsNotFound(err) {
			log.Error(err, "unable to get current state of secret")
			secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
			secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
			return ctrl.Result{}, ignoreNotFoundError(err)
		}

		eq := reflect.DeepEqual(desiredState, currentState)
		if !eq {
			log.Info("secret must be updated")
			if err := r.upsertSecret(sDef, desiredState); err != nil {
				log.Error(err, "unable to upsert secret")
				secretSyncErrorsTotal.WithLabelValues(secretNamespace, secretName).Inc()
				secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(0.0)
				return ctrl.Result{}, err
			}
			log.Info("secret updated")
			secretLastSyncStatus.WithLabelValues(secretNamespace, secretName).Set(1.0)
		}

		return ctrl.Result{RequeueAfter: r.ReconciliationPeriod}, nil

	} else {
		// SecretDefinition has been marked for deletion and contains finalizer
		if containsString(sDef.ObjectMeta.Finalizers, finalizerName) {
			if err = r.deleteSecret(secretNamespace, secretName); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "unable to delete secret")
				return ctrl.Result{}, ignoreNotFoundError(err)
			}
			log.Info("secret deleted successfully")
			// If success remove finalizer
			sDef.ObjectMeta.Finalizers = removeString(sDef.ObjectMeta.Finalizers, finalizerName)
			if err = r.Update(r.Ctx, sDef); err != nil {
				log.Error(err, "unable to remove finalizer from SecretDefinition", "finalizer", finalizerName)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
}

// SetupWithManager will register the controller
func (r *SecretDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&smv1alpha1.SecretDefinition{}).
		Complete(r)
}

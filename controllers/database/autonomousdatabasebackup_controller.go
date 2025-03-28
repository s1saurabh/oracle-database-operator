/*
** Copyright (c) 2022 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dbv4 "github.com/oracle/oracle-database-operator/apis/database/v4"
	adbfamily "github.com/oracle/oracle-database-operator/commons/adb_family"
	"github.com/oracle/oracle-database-operator/commons/oci"
)

// AutonomousDatabaseBackupReconciler reconciles a AutonomousDatabaseBackup object
type AutonomousDatabaseBackupReconciler struct {
	KubeClient client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Recorder   record.EventRecorder

	dbService oci.DatabaseService
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutonomousDatabaseBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv4.AutonomousDatabaseBackup{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 100}). // ReconcileHandler is never invoked concurrently with the same object.
		Complete(r)
}

//+kubebuilder:rbac:groups=database.oracle.com,resources=autonomousdatabasebackups,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups=database.oracle.com,resources=autonomousdatabasebackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database.oracle.com,resources=autonomousdatabases,verbs=get;list

func (r *AutonomousDatabaseBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Namespace/Name", req.NamespacedName)

	backup := &dbv4.AutonomousDatabaseBackup{}
	if err := r.KubeClient.Get(context.TODO(), req.NamespacedName, backup); err != nil {
		// Ignore not-found errors, since they can't be fixed by an immediate requeue.
		// No need to change the since we don't know if we obtain the object.
		if apiErrors.IsNotFound(err) {
			return emptyResult, nil
		}
		// Failed to get AutonomousDatabaseBackup, so we don't need to update the status
		return emptyResult, err
	}

	/******************************************************************
	* Look up the owner AutonomousDatabase and set the ownerReference
	* if the owner hasn't been set yet.
	******************************************************************/
	adbOCID, err := r.verifyTargetAdb(backup)
	if err != nil {
		return r.manageError(backup, err)
	}

	/******************************************************************
	* Get OCI database client
	******************************************************************/
	if err := r.setupOCIClients(backup); err != nil {
		return r.manageError(backup, err)
	}

	logger.Info("OCI clients configured succesfully")

	/******************************************************************
	* Get status from OCI AutonomousDatabaseBackup
	******************************************************************/
	if backup.Spec.AutonomousDatabaseBackupOCID != nil {
		backupResp, err := r.dbService.GetAutonomousDatabaseBackup(*backup.Spec.AutonomousDatabaseBackupOCID)
		if err != nil {
			return r.manageError(backup, err)
		}

		adbResp, err := r.dbService.GetAutonomousDatabase(*backupResp.AutonomousDatabaseId)
		if err != nil {
			return r.manageError(backup, err)
		}

		backup.UpdateStatusFromOCIBackup(backupResp.AutonomousDatabaseBackup, adbResp.AutonomousDatabase)
	}

	/******************************************************************
	* Requeue if the Backup is in an intermediate state
	* No-op if the Autonomous Database OCID is nil
	* To get the latest status, execute before all the reconcile logic
	******************************************************************/
	if dbv4.IsBackupIntermediateState(backup.Status.LifecycleState) {
		logger.WithName("IsIntermediateState").Info("Current lifecycleState is " + string(backup.Status.LifecycleState) + "; reconcile queued")
		return requeueResult, nil
	}

	/******************************************************************
	 * If the spec.autonomousDatabaseBackupOCID is empty.
	 * Otherwise, bind to an exisiting backup.
	 ******************************************************************/
	if backup.Spec.AutonomousDatabaseBackupOCID == nil {
		// Create a new backup
		logger.Info("Sending CreateAutonomousDatabaseBackup request to OCI")
		backupResp, err := r.dbService.CreateAutonomousDatabaseBackup(backup, adbOCID)
		if err != nil {
			return r.manageError(backup, err)
		}

		// After the creation, update the status first
		adbResp, err := r.dbService.GetAutonomousDatabase(*backupResp.AutonomousDatabaseId)
		if err != nil {
			return r.manageError(backup, err)
		}

		backup.UpdateStatusFromOCIBackup(backupResp.AutonomousDatabaseBackup, adbResp.AutonomousDatabase)
		if err := r.KubeClient.Status().Update(context.TODO(), backup); err != nil {
			return r.manageError(backup, err)
		}

		// Then update the OCID
		backup.Spec.AutonomousDatabaseBackupOCID = backupResp.Id
		backup.UpdateStatusFromOCIBackup(backupResp.AutonomousDatabaseBackup, adbResp.AutonomousDatabase)

		if err := r.KubeClient.Update(context.TODO(), backup); err != nil {
			// Do no requeue otherwise it will create multiple backups
			r.Recorder.Event(backup, corev1.EventTypeWarning, "ReconcileFailed", err.Error())
			logger.Error(err, "cannot update AutonomousDatabaseBackupOCID; stop reconcile", "AutonomousDatabaseBackupOCID", *backupResp.Id)

			return emptyResult, nil
		}

		logger.Info("AutonomousDatabaseBackupOCID updated")
		return emptyResult, nil
	}

	/******************************************************************
	*	Update the status and requeue if it's in an intermediate state
	******************************************************************/
	if err := r.KubeClient.Status().Update(context.TODO(), backup); err != nil {
		return r.manageError(backup, err)
	}

	if dbv4.IsBackupIntermediateState(backup.Status.LifecycleState) {
		logger.WithName("IsIntermediateState").Info("Reconcile queued")
		return requeueResult, nil
	}

	logger.Info("AutonomousDatabaseBackup reconciles successfully")

	return emptyResult, nil
}

// setOwnerAutonomousDatabase sets the owner of the AutonomousDatabaseBackup if the AutonomousDatabase resource with the same database OCID is found
func (r *AutonomousDatabaseBackupReconciler) setOwnerAutonomousDatabase(backup *dbv4.AutonomousDatabaseBackup, adb *dbv4.AutonomousDatabase) error {
	logger := r.Log.WithName("set-owner-reference")

	controllerutil.SetOwnerReference(adb, backup, r.Scheme)
	if err := r.KubeClient.Update(context.TODO(), backup); err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Set the owner of AutonomousDatabaseBackup %s to AutonomousDatabase %s", backup.Name, adb.Name))

	return nil
}

// verifyTargetAdb searches if the target AutonomousDatabase is in the cluster, and set the owner reference to that AutonomousDatabase if it exists.
// The function returns the OCID of the target AutonomousDatabase.
func (r *AutonomousDatabaseBackupReconciler) verifyTargetAdb(backup *dbv4.AutonomousDatabaseBackup) (string, error) {
	// Get the target ADB OCID and the ADB resource
	ownerAdb, err := adbfamily.VerifyTargetAdb(r.KubeClient, backup.Spec.Target, backup.Namespace)

	if err != nil {
		return "", err
	}

	// Set the owner reference if needed
	if len(backup.GetOwnerReferences()) == 0 && ownerAdb != nil {
		if err := r.setOwnerAutonomousDatabase(backup, ownerAdb); err != nil {
			return "", err
		}
	}

	if backup.Spec.Target.OciAdb.OCID != nil {
		return *backup.Spec.Target.OciAdb.OCID, nil
	}
	if ownerAdb != nil && ownerAdb.Spec.Details.Id != nil {
		return *ownerAdb.Spec.Details.Id, nil
	}

	return "", errors.New("cannot get the OCID of the target AutonomousDatabase")
}

func (r *AutonomousDatabaseBackupReconciler) setupOCIClients(backup *dbv4.AutonomousDatabaseBackup) error {
	var err error

	authData := oci.ApiKeyAuth{
		ConfigMapName: backup.Spec.OCIConfig.ConfigMapName,
		SecretName:    backup.Spec.OCIConfig.SecretName,
		Namespace:     backup.GetNamespace(),
	}

	provider, err := oci.GetOciProvider(r.KubeClient, authData)
	if err != nil {
		return err
	}

	r.dbService, err = oci.NewDatabaseService(r.Log, r.KubeClient, provider)
	if err != nil {
		return err
	}

	return nil
}

func (r *AutonomousDatabaseBackupReconciler) manageError(backup *dbv4.AutonomousDatabaseBackup, issue error) (ctrl.Result, error) {
	// Send event
	r.Recorder.Event(backup, corev1.EventTypeWarning, "ReconcileFailed", issue.Error())

	return emptyResult, issue
}

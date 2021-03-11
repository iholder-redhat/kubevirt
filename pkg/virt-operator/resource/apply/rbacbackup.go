package apply

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/cache"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/log"
)

func (r *Reconciler) backupRbac() error {

	rbac := r.clientset.RbacV1()

	// Backup existing ClusterRoles
	objects := r.stores.ClusterRoleCache.List()
	for _, obj := range objects {
		cachedCr, ok := obj.(*rbacv1.ClusterRole)
		if !ok || !needsUpdate(r.kv, r.stores.ClusterRoleCache, &cachedCr.ObjectMeta) {
			continue
		}
		imageTag, imageRegistry, id, ok := getInstallStrategyAnnotations(&cachedCr.ObjectMeta)
		if !ok {
			continue
		}

		// needs backup, so create a new object that will temporarily
		// backup this object while the update is in progress.
		cr := cachedCr.DeepCopy()
		cr.ObjectMeta = metav1.ObjectMeta{
			GenerateName: cachedCr.Name,
		}
		injectOperatorMetadata(r.kv, &cr.ObjectMeta, imageTag, imageRegistry, id, true)
		cr.Annotations[v1.EphemeralBackupObject] = string(cachedCr.UID)

		// Create backup
		r.expectations.ClusterRole.RaiseExpectations(r.kvKey, 1, 0)
		_, err := rbac.ClusterRoles().Create(context.Background(), cr, metav1.CreateOptions{})
		if err != nil {
			r.expectations.ClusterRole.LowerExpectations(r.kvKey, 1, 0)
			return fmt.Errorf("unable to create backup clusterrole %+v: %v", cr, err)
		}

		log.Log.V(2).Infof("backup clusterrole %v created", cr.GetName())
	}

	// Backup existing ClusterRoleBindings
	objects = r.stores.ClusterRoleBindingCache.List()
	for _, obj := range objects {
		cachedCrb, ok := obj.(*rbacv1.ClusterRoleBinding)
		if !ok || !needsUpdate(r.kv, r.stores.ClusterRoleBindingCache, &cachedCrb.ObjectMeta) {
			continue
		}
		imageTag, imageRegistry, id, ok := getInstallStrategyAnnotations(&cachedCrb.ObjectMeta)
		if !ok {
			continue
		}

		// needs backup, so create a new object that will temporarily
		// backup this object while the update is in progress.
		crb := cachedCrb.DeepCopy()
		crb.ObjectMeta = metav1.ObjectMeta{
			GenerateName: cachedCrb.Name,
		}
		injectOperatorMetadata(r.kv, &crb.ObjectMeta, imageTag, imageRegistry, id, true)
		crb.Annotations[v1.EphemeralBackupObject] = string(cachedCrb.UID)

		// Create backup
		r.expectations.ClusterRoleBinding.RaiseExpectations(r.kvKey, 1, 0)
		_, err := rbac.ClusterRoleBindings().Create(context.Background(), crb, metav1.CreateOptions{})
		if err != nil {
			r.expectations.ClusterRoleBinding.LowerExpectations(r.kvKey, 1, 0)
			return fmt.Errorf("unable to create backup clusterrolebinding %+v: %v", crb, err)
		}
		log.Log.V(2).Infof("backup clusterrolebinding %v created", crb.GetName())
	}

	// Backup existing Roles
	objects = r.stores.RoleCache.List()
	for _, obj := range objects {
		cachedCr, ok := obj.(*rbacv1.Role)
		if !ok || !needsUpdate(r.kv, r.stores.RoleCache, &cachedCr.ObjectMeta) {
			continue
		}
		imageTag, imageRegistry, id, ok := getInstallStrategyAnnotations(&cachedCr.ObjectMeta)
		if !ok {
			continue
		}

		// needs backup, so create a new object that will temporarily
		// backup this object while the update is in progress.
		cr := cachedCr.DeepCopy()
		cr.ObjectMeta = metav1.ObjectMeta{
			GenerateName: cachedCr.Name,
		}
		injectOperatorMetadata(r.kv, &cr.ObjectMeta, imageTag, imageRegistry, id, true)
		cr.Annotations[v1.EphemeralBackupObject] = string(cachedCr.UID)

		// Create backup
		r.expectations.Role.RaiseExpectations(r.kvKey, 1, 0)
		_, err := rbac.Roles(cachedCr.Namespace).Create(context.Background(), cr, metav1.CreateOptions{})
		if err != nil {
			r.expectations.Role.LowerExpectations(r.kvKey, 1, 0)
			return fmt.Errorf("unable to create backup role %+v: %v", r, err)
		}
		log.Log.V(2).Infof("backup role %v created", cr.GetName())
	}

	// Backup existing RoleBindings
	objects = r.stores.RoleBindingCache.List()
	for _, obj := range objects {
		cachedRb, ok := obj.(*rbacv1.RoleBinding)
		if !ok || !needsUpdate(r.kv, r.stores.RoleBindingCache, &cachedRb.ObjectMeta) {
			continue
		}
		imageTag, imageRegistry, id, ok := getInstallStrategyAnnotations(&cachedRb.ObjectMeta)
		if ok {
			continue
		}

		// needs backup, so create a new object that will temporarily
		// backup this object while the update is in progress.
		rb := cachedRb.DeepCopy()
		rb.ObjectMeta = metav1.ObjectMeta{
			GenerateName: cachedRb.Name,
		}
		injectOperatorMetadata(r.kv, &rb.ObjectMeta, imageTag, imageRegistry, id, true)
		rb.Annotations[v1.EphemeralBackupObject] = string(cachedRb.UID)

		// Create backup
		r.expectations.RoleBinding.RaiseExpectations(r.kvKey, 1, 0)
		_, err := rbac.RoleBindings(cachedRb.Namespace).Create(context.Background(), rb, metav1.CreateOptions{})
		if err != nil {
			r.expectations.RoleBinding.LowerExpectations(r.kvKey, 1, 0)
			return fmt.Errorf("unable to create backup rolebinding %+v: %v", rb, err)
		}
		log.Log.V(2).Infof("backup rolebinding %v created", rb.GetName())
	}

	return nil
}

func shouldBackupRBACObject(kv *v1.KubeVirt, objectMeta *metav1.ObjectMeta) (imageTag, imageRegistry, id string, shouldBackup bool) {
	curVersion, curImageRegistry, curID := getTargetVersionRegistryID(kv)
	shouldBackup = false

	if objectMatchesVersion(objectMeta, curVersion, curImageRegistry, curID, kv.GetGeneration()) {
		// matches current target version already, so doesn't need backup
		return
	}

	if objectMeta.Annotations == nil {
		return
	}

	_, ok := objectMeta.Annotations[v1.EphemeralBackupObject]
	if ok {
		// ephemeral backup objects don't need to be backed up because
		// they are the backup
		return
	}

	imageTag, imageRegistry, id, shouldBackup = getInstallStrategyAnnotations(objectMeta)

	return

}

func getInstallStrategyAnnotations(meta *metav1.ObjectMeta) (imageTag, imageRegistry, id string, ok bool) {
	ok = false

	imageTag, ok = meta.Annotations[v1.InstallStrategyVersionAnnotation]
	if !ok {
		return
	}
	imageRegistry, ok = meta.Annotations[v1.InstallStrategyRegistryAnnotation]
	if !ok {
		return
	}
	id, ok = meta.Annotations[v1.InstallStrategyIdentifierAnnotation]
	if !ok {
		return
	}

	ok = true
	return
}

func needsUpdate(kv *v1.KubeVirt, cache cache.Store, meta *metav1.ObjectMeta) bool {
	imageTag, imageRegistry, id, shouldBackup := shouldBackupRBACObject(kv, meta)
	if !shouldBackup {
		return false
	}

	// loop through cache and determine if there's an ephemeral backup
	// for this object already
	objects := cache.List()
	for _, obj := range objects {
		cachedObj, ok := obj.(*metav1.ObjectMeta)

		if !ok ||
			cachedObj.DeletionTimestamp != nil ||
			meta.Annotations == nil {
			continue
		}

		uid, ok := cachedObj.Annotations[v1.EphemeralBackupObject]
		if !ok {
			// this is not an ephemeral backup object
			continue
		}

		if uid == string(meta.UID) && objectMatchesVersion(cachedObj, imageTag, imageRegistry, id, kv.GetGeneration()) {
			// found backup. UID matches and versions match
			// note, it's possible for a single UID to have multiple backups with
			// different versions
			return false
		}
	}

	return true
}

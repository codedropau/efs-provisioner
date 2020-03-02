package provisioner

import corev1 "k8s.io/api/core/v1"

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
// @todo, Tag FileSystem as "ready for removal"
// @todo, Tag FileSystem with a date to show how old it is.
func (p *Provisioner) Delete(volume *corev1.PersistentVolume) error {
	return nil
}

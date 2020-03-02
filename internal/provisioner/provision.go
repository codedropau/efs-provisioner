package provisioner

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

// TagName provide a human-readable name for provisioned resources.
const TagName = "Name"

// Provision creates a storage asset and returns a PV object representing it.
func (p *Provisioner) Provision(options controller.ProvisionOptions) (*corev1.PersistentVolume, error) {
	// This is a consistent naming pattern for provisioning our EFS objects.
	name, err := formatName(p.params.Format, options)
	if err != nil {
		return nil, err
	}

	glog.Infof("Provisioning filesystem: %s", name)

	filesystem, err := provisionFileSystem(p.client, name, p.params.Performance)
	if err != nil {
		return nil, fmt.Errorf("failed to provision filesystem: %w", err)
	}

	err = provisionMountTargets(p.client, p.params.Subnets, name, *filesystem.FileSystemId, p.params.SecurityGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to provision mount targets: %w", err)
	}

	glog.Infof("Responding with persistent volume spec: %s", name)

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: *filesystem.FileSystemId,
			Annotations: map[string]string{
				// https://kubernetes.io/docs/concepts/storage/persistent-volumes
				// http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html
				MountOptionAnnotation: "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			// PersistentVolumeReclaimPolicy, AccessModes and Capacity are required fields.
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: corev1.ResourceList{
				// AWS EFS returns a "massive" file storage size when mounted. We replicate that here.
				corev1.ResourceStorage: resource.MustParse("8.0E"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: fmt.Sprintf("%s.efs.%s.amazonaws.com", *filesystem.FileSystemId, p.params.Region),
					Path:   "/",
				},
			},
		},
	}

	return pv, nil
}

// Helper function to provision a filesystem.
func provisionFileSystem(client efsiface.EFSAPI, name, performance string) (*efs.FileSystemDescription, error) {
	limiter := time.Tick(time.Second * 15)

	put := func(svc efsiface.EFSAPI, name string, performance string) (*efs.FileSystemDescription, error) {
		describe, err := svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			CreationToken: aws.String(name),
		})
		if err != nil {
			return nil, err
		}

		// We have found the filesystem! Give this back to the provisioner.
		if len(describe.FileSystems) == 1 {
			return describe.FileSystems[0], nil
		}

		// We dont hav the filesystem, lets provision it now.
		filesystem, err := svc.CreateFileSystem(&efs.CreateFileSystemInput{
			CreationToken:   aws.String(name),
			PerformanceMode: aws.String(string(performance)),
			Tags: []*efs.Tag{
				{
					Key:   aws.String(TagName),
					Value: aws.String(name),
				},
			},
		})
		if err != nil {
			return nil, err
		}

		return filesystem, nil
	}

	filesystem, err := put(client, name, performance)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %s", err)
	}

	for {
		glog.Infof("Waiting for filesystem to become ready: %s", name)

		filesystem, err = put(client, name, performance)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %s", err)
		}

		if *filesystem.LifeCycleState == efs.LifeCycleStateAvailable {
			break
		}

		<-limiter
	}

	return filesystem, nil
}

// Helper function to provision EFS mount targets for a filesystem.
func provisionMountTargets(client efsiface.EFSAPI, subnets []string, name, id, security string) error {
	limiter := time.Tick(time.Second * 15)

	put := func(svc efsiface.EFSAPI, id, subnet, security string) (*efs.MountTargetDescription, error) {
		// Check if a mount exists in this subnet.
		targets, err := svc.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			FileSystemId: aws.String(id),
		})
		if err != nil {
			return nil, err
		}

		for _, mount := range targets.MountTargets {
			// Check if we have already setup a mount point on a specific subnet.
			if *mount.SubnetId == subnet {
				return mount, nil
			}
		}

		// Create one if it does not exist.
		return svc.CreateMountTarget(&efs.CreateMountTargetInput{
			FileSystemId: aws.String(id),
			SubnetId:     aws.String(subnet),
			SecurityGroups: []*string{
				aws.String(security),
			},
		})
	}

	for _, subnet := range subnets {
		for {
			glog.Infof("Waiting for mount target (%s) to become ready for filesystem: %s", subnet, name)

			target, err := put(client, id, subnet, security)
			if err != nil {
				return fmt.Errorf("failed to create filesystem: %s", err)
			}

			if *target.LifeCycleState == efs.LifeCycleStateAvailable {
				break
			}

			<-limiter
		}
	}

	return nil
}

// Helper function for building hostname.
func formatName(format string, options controller.ProvisionOptions) (string, error) {
	var formatted bytes.Buffer

	t := template.Must(template.New("name").Parse(format))

	err := t.Execute(&formatted, options)
	if err != nil {
		return "", err
	}

	return formatted.String(), nil
}

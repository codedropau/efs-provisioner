package provisioner

import (
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/kelseyhightower/envconfig"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

// MountOptionAnnotation is the annotation on a PV object that specifies a
// comma separated list of mount options
const MountOptionAnnotation = "volume.beta.kubernetes.io/mount-options"

var _ controller.Provisioner = &Provisioner{}

// Provisioner for creating volumes.
type Provisioner struct {
	client efsiface.EFSAPI
	params Params
}

// Params required for provisioning volumes.
type Params struct {
	// Use the same region as the AWS client.
	Region        string   `envconfig:"AWS_REGION" default:"ap-southeast-2"`
	Format        string   `default:"{{ .PVC.ObjectMeta.Namespace }}-{{ .PVName }}"`
	Performance   string   `default:"generalPurpose"`
	SecurityGroup string   `required:"true"`
	Subnets       []string `required:"true"`
}

// New provisioner using environment variables for configuration.
func NewFromEnv(client efsiface.EFSAPI) (*Provisioner, error) {
	var params Params

	err := envconfig.Process("efs_provisioner", &params)
	if err != nil {
		return nil, err
	}

	provisioner := &Provisioner{
		client: client,
		params: params,
	}

	return provisioner, nil
}
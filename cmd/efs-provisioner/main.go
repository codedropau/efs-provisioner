package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"

	"github.com/codedropau/efs-provisioner/internal/provisioner"
)

func main() {
	flag.Parse()
	flag.Set("logtostderr", "true")

	k8sconfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %s", err)
	}
	k8sclient, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		glog.Fatalf("Failed to create client: %s", err)
	}

	serverVersion, err := k8sclient.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %s", err)
	}

	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		apiVersion = fmt.Sprintf("skpr.io/standard")
	}

	client := efs.New(session.New())

	provisioner, err := provisioner.NewFromEnv(client)
	if err != nil {
		glog.Fatalf("Failed to create provisioner: %s", err)
	}

	glog.Infof("Running provisioner: %s", apiVersion)

	// Start the provision controller which will dynamically provision NFS PVs
	pc := controller.NewProvisionController(
		k8sclient,
		apiVersion,
		provisioner,
		serverVersion.GitVersion,
	)

	pc.Run(wait.NeverStop)
}

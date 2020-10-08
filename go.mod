module github.com/codedropau/efs-provisioner

go 1.13

require (
	github.com/aws/aws-sdk-go v1.28.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/go-ps v0.0.0-20190716172923-621e5597135b
	github.com/previousnext/k8s-aws-efs v0.0.0-20200114055257-fb808e061ce5 // indirect
	github.com/prometheus/common v0.8.0
	github.com/shirou/gopsutil v2.19.12+incompatible
	github.com/stretchr/testify v1.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.0.0-20200113233857-bcaa73156d59
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.1+incompatible
)

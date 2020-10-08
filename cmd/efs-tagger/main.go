package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliKey   = kingpin.Arg("key", "Key of the tag to assign to all filesystems").String()
	cliValue = kingpin.Arg("value", "Value of the tag to assign to all filesystems").String()
)

func main() {
	kingpin.Parse()

	svc := efs.New(session.Must(session.NewSession()))

	input := &efs.DescribeFileSystemsInput{}

	var list []*efs.FileSystemDescription

	for {
		resp, err := svc.DescribeFileSystems(input)
		if err != nil {
			panic(err)
		}

		for _, item := range resp.FileSystems {
			list = append(list, item)
		}

		if resp.NextMarker == nil {
			break
		}

		input.Marker = resp.NextMarker
	}

	var tagless []*efs.FileSystemDescription

	for _, item := range list {
		if hasTag(item.Tags, *cliKey, *cliValue) {
			continue
		}

		tagless = append(tagless, item)
	}

	for _, item := range tagless {
		fmt.Printf("Applying tag to resource: %s (%s)\n", *item.Name, *item.FileSystemId)

		_, err := svc.TagResource(&efs.TagResourceInput{
			ResourceId: item.FileSystemId,
			Tags: []*efs.Tag{
				{
					Key: cliKey,
					Value: cliValue,
				},
			},
		})
		if err != nil {
			panic(err)
		}
	}
}

// Helper function to tag EFS volumes.
func hasTag(tags []*efs.Tag, key, value string) bool {
	for _, tag := range tags {
		if *tag.Key != key {
			continue
		}

		if *tag.Value != value {
			continue
		}

		return true
	}

	return false
}

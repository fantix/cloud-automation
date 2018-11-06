// Tutorial on writing an out-of-tree dynamic provisioner:
// https://github.com/kubernetes-incubator/nfs-provisioner/tree/master/demo/hostpath-provisioner

// Example of another goofys provisioner:
// https://github.com/previousnext/k8s-aws-goofys

// This file is modeled off of:
// https://github.com/previousnext/k8s-aws-goofys/blob/master/controller/provisioner/provisioner.go

// possible instructions on how to use this??
// https://github.com/kubernetes-incubator/nfs-provisioner

package provisioner

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/previousnext/k8s-aws-goofys/flexvolume/cmd"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AnnotationUID for overriding the default UID (www-data)
	AnnotationUID = "pv.storage.skpr.io/uid"
	// DefaultAnnotationUID defaults the www-data uid
	DefaultAnnotationUID = "33"
	// AnnotationGID for overriding the default UID (www-data)
	AnnotationGID = "pv.storage.skpr.io/gid"
	// DefaultAnnotationGID defaults the www-data gid
	DefaultAnnotationGID = "33"
	// DriverName for installing a Flexvolume
	DriverName = "gen3/goofys"
	// DriverType for providing developers with more context around this storage option
	DriverType = "fuse"
	// DefaultEnvFormat is a fallback for when a format is not provided
	DefaultEnvFormat = "goofys-{{ .PVC.ObjectMeta.Namespace }}-{{ .PVC.ObjectMeta.Name }}"
	// DefaultEnvRegion is a fallback for when a region is not provided
	DefaultEnvRegion = "us-east-1"
)

var _ controller.Provisioner = &goofys{}

type goofys struct {
	region string
	format string
}

// New returns a Goofys provisioner.
func New(region, format string) (controller.Provisioner, error) {
	if region == "" {
		region = DefaultEnvRegion
	}

	if format == "" {
		format = DefaultEnvFormat
	}

	provisioner := &goofys{
		region: region,
		format: format,
	}

	return provisioner, nil
}

// Provision creates a storage asset and returns a PV object representing it.
func (p *goofys) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	// This is a consistent naming pattern for provisioning our S3 bucket.
	name, err := formatName(p.format, options)
	if err != nil {
		return nil, err
	}

	glog.Infof("Provisioning bucket: %s", name)

	// Check if the bucket exists, if it does not, create the bucket if it does not exist.
	svc := s3.New(session.New(&aws.Config{Region: aws.String(p.region)}))
	params := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err = svc.CreateBucket(params)
	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() != http.StatusConflict {
				return nil, err
			}
		}
	}

	var (
		uid = DefaultAnnotationUID
		gid = DefaultAnnotationGID
	)

	if val, ok := options.PVC.ObjectMeta.Annotations[AnnotationUID]; ok {
		uid = val
	}

	if val, ok := options.PVC.ObjectMeta.Annotations[AnnotationGID]; ok {
		gid = val
	}

	glog.Infof("Responding with persistent volume spec: %s", name)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("8.0E"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexPersistentVolumeSource{
					Driver: DriverName,
					FSType: DriverType,
					Options: map[string]string{
						cmd.OptionBucket: name //,
						//cmd.OptionUID:    uid,
						//cmd.OptionGID:    gid,
					},
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
// @todo, Tag Bucket as "ready for removal"
// @todo, Tag Bucket with a date to show how old it is.
func (p *goofys) Delete(volume *v1.PersistentVolume) error {
	return nil
}

// Helper function for building hostname.
func formatName(format string, options controller.VolumeOptions) (string, error) {
	var formatted bytes.Buffer

	t := template.Must(template.New("name").Parse(format))

	err := t.Execute(&formatted, options)
	if err != nil {
		return "", err
	}

	return formatted.String(), nil
}
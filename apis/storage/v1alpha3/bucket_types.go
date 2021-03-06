/*
Copyright 2019 The Crossplane Authors.

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

package v1alpha3

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// A LocalPermissionType is a type of permission that may be granted to a
// Bucket.
type LocalPermissionType string

const (
	// ReadOnlyPermission will grant read objects in a bucket
	ReadOnlyPermission LocalPermissionType = "Read"
	// WriteOnlyPermission will grant write/delete objects in a bucket
	WriteOnlyPermission LocalPermissionType = "Write"
	// ReadWritePermission LocalPermissionType Grant both read and write permissions
	ReadWritePermission LocalPermissionType = "ReadWrite"
)

//Tag is a metadata assigned to an Amazon S3 Bucket consisting of a key-value pair.
// Please also see https://docs.aws.amazon.com/AmazonS3/latest/API/API_Tag.html
type Tag struct {
	// Name of the object key
	Key string `json:"key"`

	// Value of the tag
	Value string `json:"value"`
}

// S3BucketParameters define the desired state of an AWS S3 Bucket.
type S3BucketParameters struct {
	// Region of the bucket.
	Region string `json:"region"`

	// CannedACL applies a standard AWS built-in ACL for common bucket use
	// cases.
	// +kubebuilder:validation:Enum=private;public-read;public-read-write;authenticated-read;log-delivery-write;aws-exec-read
	// +optional
	CannedACL *s3.BucketCannedACL `json:"cannedACL,omitempty"`

	// Versioning enables versioning of objects stored in this bucket.
	// +optional
	Versioning bool `json:"versioning,omitempty"`

	// IAMUsername is the name of an IAM user that is automatically created and
	// granted access to this bucket by Crossplane at bucket creation time.
	IAMUsername string `json:"iamUsername,omitempty"`

	// LocalPermission is the permissions granted on the bucket for the provider
	// specific bucket service account that is available in a secret after
	// provisioning.
	// +kubebuilder:validation:Enum=Read;Write;ReadWrite
	LocalPermission *LocalPermissionType `json:"localPermission"`

	// A list of key-value pairs to label the S3 Bucket
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// S3BucketSpec defines the desired state of S3Bucket
type S3BucketSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	S3BucketParameters           `json:",inline"`
}

// S3BucketStatus defines the observed state of S3Bucket
type S3BucketStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// ProviderID is the AWS identifier for this bucket.
	ProviderID string `json:"providerID,omitempty"`

	// LastUserPolicyVersion is the most recent version of the policy associated
	// with this bucket's IAMUser.
	LastUserPolicyVersion int `json:"lastUserPolicyVersion,omitempty"`

	// LastLocalPermission is the most recent local permission that was set for
	// this bucket.
	LastLocalPermission LocalPermissionType `json:"lastLocalPermission,omitempty"`
}

// +kubebuilder:object:root=true

// An S3Bucket is a managed resource that represents an AWS S3 Bucket.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="PREDEFINED-ACL",type="string",JSONPath=".spec.cannedACL"
// +kubebuilder:printcolumn:name="LOCAL-PERMISSION",type="string",JSONPath=".spec.localPermission"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type S3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketSpec   `json:"spec"`
	Status S3BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// S3BucketList contains a list of S3Buckets
type S3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Bucket `json:"items"`
}

// SetUserPolicyVersion specifies this bucket's policy version.
func (b *S3Bucket) SetUserPolicyVersion(policyVersion string) error {
	policyInt, err := strconv.Atoi(policyVersion[1:])
	if err != nil {
		return err
	}
	b.Status.LastUserPolicyVersion = policyInt
	b.Status.LastLocalPermission = *b.Spec.LocalPermission

	return nil
}

// HasPolicyChanged returns true if the bucket's policy is older than the
// supplied version.
func (b *S3Bucket) HasPolicyChanged(policyVersion string) (bool, error) {
	if *b.Spec.LocalPermission != b.Status.LastLocalPermission {
		return true, nil
	}
	policyInt, err := strconv.Atoi(policyVersion[1:])
	if err != nil {
		return false, err
	}
	if b.Status.LastUserPolicyVersion != policyInt && b.Status.LastUserPolicyVersion < policyInt {
		return true, nil
	}

	return false, nil
}

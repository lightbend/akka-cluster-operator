// +build !ignore_autogenerated

// Code generated by operator-sdk-v0.15.2-x86_64-apple-darwin. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaCluster) DeepCopyInto(out *AkkaCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(AkkaClusterStatus)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaCluster.
func (in *AkkaCluster) DeepCopy() *AkkaCluster {
	if in == nil {
		return nil
	}
	out := new(AkkaCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AkkaCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterList) DeepCopyInto(out *AkkaClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AkkaCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterList.
func (in *AkkaClusterList) DeepCopy() *AkkaClusterList {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AkkaClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterManagementStatus) DeepCopyInto(out *AkkaClusterManagementStatus) {
	*out = *in
	if in.Members != nil {
		in, out := &in.Members, &out.Members
		*out = make([]AkkaClusterMemberStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Unreachable != nil {
		in, out := &in.Unreachable, &out.Unreachable
		*out = make([]AkkaClusterUnreachableMemberStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.OldestPerRole != nil {
		in, out := &in.OldestPerRole, &out.OldestPerRole
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterManagementStatus.
func (in *AkkaClusterManagementStatus) DeepCopy() *AkkaClusterManagementStatus {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterManagementStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterMemberStatus) DeepCopyInto(out *AkkaClusterMemberStatus) {
	*out = *in
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterMemberStatus.
func (in *AkkaClusterMemberStatus) DeepCopy() *AkkaClusterMemberStatus {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterMemberStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterSpec) DeepCopyInto(out *AkkaClusterSpec) {
	*out = *in
	in.DeploymentSpec.DeepCopyInto(&out.DeploymentSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterSpec.
func (in *AkkaClusterSpec) DeepCopy() *AkkaClusterSpec {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterStatus) DeepCopyInto(out *AkkaClusterStatus) {
	*out = *in
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
	in.Cluster.DeepCopyInto(&out.Cluster)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterStatus.
func (in *AkkaClusterStatus) DeepCopy() *AkkaClusterStatus {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkkaClusterUnreachableMemberStatus) DeepCopyInto(out *AkkaClusterUnreachableMemberStatus) {
	*out = *in
	if in.ObservedBy != nil {
		in, out := &in.ObservedBy, &out.ObservedBy
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkkaClusterUnreachableMemberStatus.
func (in *AkkaClusterUnreachableMemberStatus) DeepCopy() *AkkaClusterUnreachableMemberStatus {
	if in == nil {
		return nil
	}
	out := new(AkkaClusterUnreachableMemberStatus)
	in.DeepCopyInto(out)
	return out
}

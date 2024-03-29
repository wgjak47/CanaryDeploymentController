// +build !ignore_autogenerated

/*

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CanaryDeployment) DeepCopyInto(out *CanaryDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CanaryDeployment.
func (in *CanaryDeployment) DeepCopy() *CanaryDeployment {
	if in == nil {
		return nil
	}
	out := new(CanaryDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CanaryDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CanaryDeploymentList) DeepCopyInto(out *CanaryDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CanaryDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CanaryDeploymentList.
func (in *CanaryDeploymentList) DeepCopy() *CanaryDeploymentList {
	if in == nil {
		return nil
	}
	out := new(CanaryDeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CanaryDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CanaryDeploymentSpec) DeepCopyInto(out *CanaryDeploymentSpec) {
	*out = *in
	in.CommonSpec.DeepCopyInto(&out.CommonSpec)
	if in.AppSpecs != nil {
		in, out := &in.AppSpecs, &out.AppSpecs
		*out = make(map[string]DeploymentPatchSpec, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CanaryDeploymentSpec.
func (in *CanaryDeploymentSpec) DeepCopy() *CanaryDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(CanaryDeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CanaryDeploymentStatus) DeepCopyInto(out *CanaryDeploymentStatus) {
	*out = *in
	if in.AppStatus != nil {
		in, out := &in.AppStatus, &out.AppStatus
		*out = make(map[string]appsv1.DeploymentStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CanaryDeploymentStatus.
func (in *CanaryDeploymentStatus) DeepCopy() *CanaryDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(CanaryDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentPatchSpec) DeepCopyInto(out *DeploymentPatchSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentPatchSpec.
func (in *DeploymentPatchSpec) DeepCopy() *DeploymentPatchSpec {
	if in == nil {
		return nil
	}
	out := new(DeploymentPatchSpec)
	in.DeepCopyInto(out)
	return out
}

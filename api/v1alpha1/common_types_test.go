/*
Copyright 2021.

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

package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestImageSpec_Default(t *testing.T) {
	type fields struct {
		Name           *string
		Tag            *string
		PullSecretName *string
		PullPolicy     *corev1.PullPolicy
	}

	type args struct {
		def defaultImageSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ImageSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
				PullPolicy: func() *corev1.PullPolicy {
					p := corev1.PullIfNotPresent
					return &p
				}(),
			}},
			want: &ImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
				PullPolicy: func() *corev1.PullPolicy {
					p := corev1.PullIfNotPresent
					return &p
				}(),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				Name: ptr.To("explicit"),
				PullPolicy: func() *corev1.PullPolicy {
					p := corev1.PullAlways
					return &p
				}(),
			},
			args: args{def: defaultImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
				PullPolicy: func() *corev1.PullPolicy {
					p := corev1.PullIfNotPresent
					return &p
				}(),
			}},
			want: &ImageSpec{
				Name:           ptr.To("explicit"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
				PullPolicy: func() *corev1.PullPolicy {
					p := corev1.PullAlways
					return &p
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ImageSpec{
				Name:           tt.fields.Name,
				Tag:            tt.fields.Tag,
				PullSecretName: tt.fields.PullSecretName,
				PullPolicy:     tt.fields.PullPolicy,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("ImageSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestImageSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *ImageSpec
		want bool
	}{
		{"Wants false if empty", &ImageSpec{}, false},
		{"Wants false if nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("ImageSpec.IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializeImageSpec(t *testing.T) {
	type args struct {
		spec *ImageSpec
		def  defaultImageSpec
	}

	tests := []struct {
		name string
		args args
		want *ImageSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
			}},
			want: &ImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
			},
		},
		{
			name: "Initializes the struct with appropriate defaults if empty",
			args: args{&ImageSpec{}, defaultImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
			}},
			want: &ImageSpec{
				Name:           ptr.To("name"),
				Tag:            ptr.To("tag"),
				PullSecretName: ptr.To("pullSecret"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeImageSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeImageSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProbeSpec_Default(t *testing.T) {
	type fields struct {
		InitialDelaySeconds *int32
		TimeoutSeconds      *int32
		PeriodSeconds       *int32
		SuccessThreshold    *int32
		FailureThreshold    *int32
	}

	type args struct {
		def defaultProbeSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ProbeSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultProbeSpec{
				InitialDelaySeconds: ptr.To[int32](1),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			}},
			want: &ProbeSpec{
				InitialDelaySeconds: ptr.To[int32](1),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				InitialDelaySeconds: ptr.To[int32](9999),
			},
			args: args{def: defaultProbeSpec{
				InitialDelaySeconds: ptr.To[int32](1),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			}},
			want: &ProbeSpec{
				InitialDelaySeconds: ptr.To[int32](9999),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ProbeSpec{
				InitialDelaySeconds: tt.fields.InitialDelaySeconds,
				TimeoutSeconds:      tt.fields.TimeoutSeconds,
				PeriodSeconds:       tt.fields.PeriodSeconds,
				SuccessThreshold:    tt.fields.SuccessThreshold,
				FailureThreshold:    tt.fields.FailureThreshold,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("ProbeSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestProbeSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *ProbeSpec
		want bool
	}{
		{"Wants true if empty", &ProbeSpec{}, true},
		{"Wants false if nil", nil, false},
		{"Wants false if other", &ProbeSpec{InitialDelaySeconds: ptr.To[int32](1)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("ProbeSpec.IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializeProbeSpec(t *testing.T) {
	type args struct {
		spec *ProbeSpec
		def  defaultProbeSpec
	}

	tests := []struct {
		name string
		args args
		want *ProbeSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultProbeSpec{
				InitialDelaySeconds: ptr.To[int32](1),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			}},
			want: &ProbeSpec{
				InitialDelaySeconds: ptr.To[int32](1),
				TimeoutSeconds:      ptr.To[int32](2),
				PeriodSeconds:       ptr.To[int32](3),
				SuccessThreshold:    ptr.To[int32](4),
				FailureThreshold:    ptr.To[int32](5),
			},
		},
		{
			name: "Deactivated",
			args: args{&ProbeSpec{}, defaultProbeSpec{}},
			want: &ProbeSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeProbeSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeProbeSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadBalancerSpec_Default(t *testing.T) {
	type fields struct {
		ProxyProtocol                           *bool
		CrossZoneLoadBalancingEnabled           *bool
		ConnectionDrainingEnabled               *bool
		ConnectionDrainingTimeout               *int32
		ConnectionHealthcheckHealthyThreshold   *int32
		ConnectionHealthcheckUnhealthyThreshold *int32
		ConnectionHealthcheckInterval           *int32
		ConnectionHealthcheckTimeout            *int32
	}

	type args struct {
		def ElasticLoadBalancerSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ElasticLoadBalancerSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			}},
			want: &ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				ProxyProtocol: ptr.To(false),
			},
			args: args{def: ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			}},
			want: &ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(false),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ElasticLoadBalancerSpec{
				ProxyProtocol:                 tt.fields.ProxyProtocol,
				CrossZoneLoadBalancingEnabled: tt.fields.CrossZoneLoadBalancingEnabled,
				ConnectionDrainingEnabled:     tt.fields.ConnectionDrainingEnabled,
				ConnectionDrainingTimeout:     tt.fields.ConnectionDrainingTimeout,
				HealthcheckHealthyThreshold:   tt.fields.ConnectionHealthcheckHealthyThreshold,
				HealthcheckUnhealthyThreshold: tt.fields.ConnectionHealthcheckUnhealthyThreshold,
				HealthcheckInterval:           tt.fields.ConnectionHealthcheckInterval,
				HealthcheckTimeout:            tt.fields.ConnectionHealthcheckTimeout,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("LoadBalancerSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestInitializeLoadBalancerSpec(t *testing.T) {
	type args struct {
		spec *ElasticLoadBalancerSpec
		def  ElasticLoadBalancerSpec
	}

	tests := []struct {
		name string
		args args
		want *ElasticLoadBalancerSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			}},
			want: &ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			},
		},
		{
			name: "Initializes the struct with appropriate defaults if empty",
			args: args{&ElasticLoadBalancerSpec{}, ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			}},
			want: &ElasticLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
				ConnectionDrainingEnabled:     ptr.To(true),
				ConnectionDrainingTimeout:     ptr.To[int32](1),
				HealthcheckHealthyThreshold:   ptr.To[int32](2),
				HealthcheckUnhealthyThreshold: ptr.To[int32](3),
				HealthcheckInterval:           ptr.To[int32](4),
				HealthcheckTimeout:            ptr.To[int32](5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeElasticLoadBalancerSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeLoadBalancerSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNLBLoadBalancerSpec_Default(t *testing.T) {
	type fields struct {
		ProxyProtocol                 *bool
		CrossZoneLoadBalancingEnabled *bool
	}

	type args struct {
		def NetworkLoadBalancerSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *NetworkLoadBalancerSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			}},
			want: &NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				ProxyProtocol: ptr.To(false),
			},
			args: args{def: NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			}},
			want: &NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(false),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &NetworkLoadBalancerSpec{
				ProxyProtocol:                 tt.fields.ProxyProtocol,
				CrossZoneLoadBalancingEnabled: tt.fields.CrossZoneLoadBalancingEnabled,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("NLBLoadBalancerSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestInitializeNLBLoadBalancerSpec(t *testing.T) {
	type args struct {
		spec *NetworkLoadBalancerSpec
		def  NetworkLoadBalancerSpec
	}

	tests := []struct {
		name string
		args args
		want *NetworkLoadBalancerSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			}},
			want: &NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			},
		},
		{
			name: "Initializes the struct with appropriate defaults if empty",
			args: args{&NetworkLoadBalancerSpec{}, NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			}},
			want: &NetworkLoadBalancerSpec{
				ProxyProtocol:                 ptr.To(true),
				CrossZoneLoadBalancingEnabled: ptr.To(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeNetworkLoadBalancerSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeNLBLoadBalancerSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrafanaDashboardSpec_Default(t *testing.T) {
	type fields struct {
		SelectorKey   *string
		SelectorValue *string
	}

	type args struct {
		def defaultGrafanaDashboardSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *GrafanaDashboardSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultGrafanaDashboardSpec{
				SelectorKey:   ptr.To("key"),
				SelectorValue: ptr.To("label"),
			}},
			want: &GrafanaDashboardSpec{
				SelectorKey:   ptr.To("key"),
				SelectorValue: ptr.To("label"),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				SelectorKey: ptr.To("xxxx"),
			},
			args: args{def: defaultGrafanaDashboardSpec{
				SelectorKey:   ptr.To("key"),
				SelectorValue: ptr.To("label"),
			}},
			want: &GrafanaDashboardSpec{
				SelectorKey:   ptr.To("xxxx"),
				SelectorValue: ptr.To("label"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &GrafanaDashboardSpec{
				SelectorKey:   tt.fields.SelectorKey,
				SelectorValue: tt.fields.SelectorValue,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("GrafanaDashboardSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestGrafanaDashboardSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *GrafanaDashboardSpec
		want bool
	}{
		{"Wants true if empty", &GrafanaDashboardSpec{}, true},
		{"Wants false if nil", nil, false},
		{"Wants false if other", &GrafanaDashboardSpec{SelectorKey: ptr.To("key")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("GrafanaDashboardSpec_IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializeGrafanaDashboardSpec(t *testing.T) {
	type args struct {
		spec *GrafanaDashboardSpec
		def  defaultGrafanaDashboardSpec
	}

	tests := []struct {
		name string
		args args
		want *GrafanaDashboardSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultGrafanaDashboardSpec{
				SelectorKey:   ptr.To("key"),
				SelectorValue: ptr.To("label"),
			}},
			want: &GrafanaDashboardSpec{
				SelectorKey:   ptr.To("key"),
				SelectorValue: ptr.To("label"),
			},
		},
		{
			name: "Deactivated",
			args: args{&GrafanaDashboardSpec{}, defaultGrafanaDashboardSpec{}},
			want: &GrafanaDashboardSpec{},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeGrafanaDashboardSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeGrafanaDashboardSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodDisruptionBudgetSpec_Default(t *testing.T) {
	type fields struct {
		MinAvailable   *intstr.IntOrString
		MaxUnavailable *intstr.IntOrString
	}

	type args struct {
		def defaultPodDisruptionBudgetSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *PodDisruptionBudgetSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultPodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("default")),
				MaxUnavailable: nil,
			}},
			want: &PodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("default")),
				MaxUnavailable: nil,
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				MinAvailable: ptr.To(intstr.FromString("explicit")),
			},
			args: args{def: defaultPodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("default")),
				MaxUnavailable: nil,
			}},
			want: &PodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("explicit")),
				MaxUnavailable: nil,
			},
		},
		{
			name: "Only one of MinAvailable or MaxUnavailable can be set",
			fields: fields{
				MinAvailable: ptr.To(intstr.FromString("explicit")),
			},
			args: args{def: defaultPodDisruptionBudgetSpec{
				MinAvailable:   nil,
				MaxUnavailable: ptr.To(intstr.FromString("default")),
			}},
			want: &PodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("explicit")),
				MaxUnavailable: nil,
			},
		},
		{
			name:   "Only one of MinAvailable or MaxUnavailable can be set (II)",
			fields: fields{},
			args: args{def: defaultPodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.IntOrString{Type: intstr.String, StrVal: "defaultMin"}),
				MaxUnavailable: ptr.To(intstr.IntOrString{Type: intstr.String, StrVal: "defaultMax"}),
			}},
			want: &PodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.IntOrString{Type: intstr.String, StrVal: "defaultMin"}),
				MaxUnavailable: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &PodDisruptionBudgetSpec{
				MinAvailable:   tt.fields.MinAvailable,
				MaxUnavailable: tt.fields.MaxUnavailable,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("PodDisruptionBudgetSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestPodDisruptionBudgetSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *PodDisruptionBudgetSpec
		want bool
	}{
		{"Wants true if empty", &PodDisruptionBudgetSpec{}, true},
		{"Wants false if nil", nil, false},
		{"Wants false if other", &PodDisruptionBudgetSpec{MinAvailable: ptr.To(intstr.FromInt(1))}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("PodDisruptionBudgetSpec.IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializePodDisruptionBudgetSpec(t *testing.T) {
	type args struct {
		spec *PodDisruptionBudgetSpec
		def  defaultPodDisruptionBudgetSpec
	}

	tests := []struct {
		name string
		args args
		want *PodDisruptionBudgetSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultPodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("default")),
				MaxUnavailable: nil,
			}},
			want: &PodDisruptionBudgetSpec{
				MinAvailable:   ptr.To(intstr.FromString("default")),
				MaxUnavailable: nil,
			},
		},
		{
			name: "Deactivated",
			args: args{&PodDisruptionBudgetSpec{}, defaultPodDisruptionBudgetSpec{}},
			want: &PodDisruptionBudgetSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializePodDisruptionBudgetSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializePodDisruptionBudgetSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHorizontalPodAutoscalerSpec_Default(t *testing.T) {
	type fields struct {
		MinReplicas         *int32
		MaxReplicas         *int32
		ResourceName        *string
		ResourceUtilization *int32
	}

	type args struct {
		def defaultHorizontalPodAutoscalerSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HorizontalPodAutoscalerSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultHorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](1),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			}},
			want: &HorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](1),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				MinReplicas: ptr.To[int32](9999),
			},
			args: args{def: defaultHorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](1),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			}},
			want: &HorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](9999),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &HorizontalPodAutoscalerSpec{
				MinReplicas:         tt.fields.MinReplicas,
				MaxReplicas:         tt.fields.MaxReplicas,
				ResourceName:        tt.fields.ResourceName,
				ResourceUtilization: tt.fields.ResourceUtilization,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("HorizontalPodAutoscalerSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestHorizontalPodAutoscalerSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *HorizontalPodAutoscalerSpec
		want bool
	}{
		{"Wants true if empty", &HorizontalPodAutoscalerSpec{}, true},
		{"Wants false if nil", nil, false},
		{"Wants false if other", &HorizontalPodAutoscalerSpec{MinReplicas: ptr.To[int32](1)}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("HorizontalPodAutoscalerSpec.IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializeHorizontalPodAutoscalerSpec(t *testing.T) {
	type args struct {
		spec *HorizontalPodAutoscalerSpec
		def  defaultHorizontalPodAutoscalerSpec
	}

	tests := []struct {
		name string
		args args
		want *HorizontalPodAutoscalerSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultHorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](1),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			}},
			want: &HorizontalPodAutoscalerSpec{
				MinReplicas:         ptr.To[int32](1),
				MaxReplicas:         ptr.To[int32](2),
				ResourceUtilization: ptr.To[int32](3),
				ResourceName:        ptr.To("xxxx"),
			},
		},
		{
			name: "Deactivated",
			args: args{&HorizontalPodAutoscalerSpec{}, defaultHorizontalPodAutoscalerSpec{}},
			want: &HorizontalPodAutoscalerSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeHorizontalPodAutoscalerSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeHorizontalPodAutoscalerSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRequirementsSpec_Default(t *testing.T) {
	type fields struct {
		Limits   corev1.ResourceList
		Requests corev1.ResourceList
	}

	type args struct {
		def defaultResourceRequirementsSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ResourceRequirementsSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			}},
			want: &ResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				}},
			args: args{def: defaultResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			}},
			want: &ResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ResourceRequirementsSpec{
				Limits:   tt.fields.Limits,
				Requests: tt.fields.Requests,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("ResourceRequirementsSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestResourceRequirementsSpec_IsDeactivated(t *testing.T) {
	tests := []struct {
		name string
		spec *ResourceRequirementsSpec
		want bool
	}{
		{"Wants true if empty", &ResourceRequirementsSpec{}, true},
		{"Wants false if nil", nil, false},
		{"Wants false if other",
			&ResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsDeactivated(); got != tt.want {
				t.Errorf("ResourceRequirementsSpec.IsDeactivated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitializeResourceRequirementsSpec(t *testing.T) {
	type args struct {
		spec *ResourceRequirementsSpec
		def  defaultResourceRequirementsSpec
	}

	tests := []struct {
		name string
		args args
		want *ResourceRequirementsSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				},
			}},
			want: &ResourceRequirementsSpec{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				},
			},
		},
		{
			name: "Deactivated",
			args: args{&ResourceRequirementsSpec{}, defaultResourceRequirementsSpec{}},
			want: &ResourceRequirementsSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeResourceRequirementsSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeResourceRequirementsSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringOrDefault(t *testing.T) {
	type args struct {
		value    *string
		defValue *string
	}

	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "Value explicitly set",
			args: args{
				value:    ptr.To("value"),
				defValue: ptr.To("default"),
			},
			want: ptr.To("value"),
		},
		{
			name: "Value not set",
			args: args{
				value:    nil,
				defValue: ptr.To("default"),
			},
			want: ptr.To("default"),
		},
		{
			name: "Nor value not default set",
			args: args{
				value:    nil,
				defValue: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringOrDefault(tt.args.value, tt.args.defValue)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringOrDefault() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func Test_intOrDefault(t *testing.T) {
	type args struct {
		value    *int32
		defValue *int32
	}

	tests := []struct {
		name string
		args args
		want *int32
	}{
		{
			name: "Value explicitly set",
			args: args{
				value:    ptr.To[int32](100),
				defValue: ptr.To[int32](10),
			},
			want: ptr.To[int32](100),
		},
		{
			name: "Value not set",
			args: args{
				value:    nil,
				defValue: ptr.To[int32](10),
			},
			want: ptr.To[int32](10),
		},
		{
			name: "Nor value not default set",
			args: args{
				value:    nil,
				defValue: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intOrDefault(tt.args.value, tt.args.defValue)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("intOrDefault() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func Test_boolOrDefault(t *testing.T) {
	type args struct {
		value    *bool
		defValue *bool
	}

	tests := []struct {
		name string
		args args
		want *bool
	}{
		{
			name: "Value explicitly set",
			args: args{
				value:    ptr.To(true),
				defValue: ptr.To(false),
			},
			want: ptr.To(true),
		},
		{
			name: "Value not set",
			args: args{
				value:    nil,
				defValue: ptr.To(false),
			},
			want: ptr.To(false),
		},
		{
			name: "Nor value not default set",
			args: args{
				value:    nil,
				defValue: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolOrDefault(tt.args.value, tt.args.defValue)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("boolOrDefault() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestCanary_CanarySpec(t *testing.T) {
	type fields struct {
		ImageName *string
		ImageTag  *string
		Replicas  *int32
		Patches   []string
	}

	type args struct {
		spec       any
		canarySpec any
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "Returns a canary spec",
			fields: fields{
				Patches: []string{
					`[{"op": "replace", "path": "/image/name", "value": "new"}]`,
				},
			},
			args: args{
				spec: &BackendSpec{
					Image: &ImageSpec{
						Name: ptr.To("old"),
						Tag:  ptr.To("tag"),
					},
				},
				canarySpec: &BackendSpec{},
			},
			want: &BackendSpec{
				Image: &ImageSpec{
					Name: ptr.To("new"),
					Tag:  ptr.To("tag"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Canary{
				ImageName: tt.fields.ImageName,
				ImageTag:  tt.fields.ImageTag,
				Replicas:  tt.fields.Replicas,
				Patches:   tt.fields.Patches,
			}

			err := c.PatchSpec(tt.args.spec, tt.args.canarySpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Canary.CanarySpec() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if diff := deep.Equal(tt.args.canarySpec, tt.want); len(diff) > 0 {
				t.Errorf("Canary.CanarySpec() = diff %v", diff)
			}
		})
	}
}

func TestExternalSecretSecretStoreReferenceSpec_Default(t *testing.T) {
	type fields struct {
		Name *string
		Kind *string
	}

	type args struct {
		def defaultExternalSecretSecretStoreReferenceSpec
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ExternalSecretSecretStoreReferenceSpec
	}{
		{
			name:   "Sets defaults",
			fields: fields{},
			args: args{def: defaultExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("vault-mgmt"),
				Kind: ptr.To("ClusterSecretStore"),
			}},
			want: &ExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("vault-mgmt"),
				Kind: ptr.To("ClusterSecretStore"),
			},
		},
		{
			name: "Combines explicitly set values with defaults",
			fields: fields{
				Name: ptr.To("other-vault"),
			},
			args: args{def: defaultExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("vault-mgmt"),
				Kind: ptr.To("ClusterSecretStore"),
			}},
			want: &ExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("other-vault"),
				Kind: ptr.To("ClusterSecretStore"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ExternalSecretSecretStoreReferenceSpec{
				Name: tt.fields.Name,
				Kind: tt.fields.Kind,
			}
			spec.Default(tt.args.def)

			if !reflect.DeepEqual(spec, tt.want) {
				t.Errorf("ExternalSecretSecretStoreReferenceSpec_Default() = %v, want %v", *spec, *tt.want)
			}
		})
	}
}

func TestInitializeVaultSecretStoreReferenceSpec(t *testing.T) {
	type args struct {
		spec *ExternalSecretSecretStoreReferenceSpec
		def  defaultExternalSecretSecretStoreReferenceSpec
	}

	tests := []struct {
		name string
		args args
		want *ExternalSecretSecretStoreReferenceSpec
	}{
		{
			name: "Initializes the struct with appropriate defaults if nil",
			args: args{nil, defaultExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("vault-mgmt"),
				Kind: ptr.To("ClusterSecretStore"),
			}},
			want: &ExternalSecretSecretStoreReferenceSpec{
				Name: ptr.To("vault-mgmt"),
				Kind: ptr.To("ClusterSecretStore"),
			},
		},
		{
			name: "Deactivated",
			args: args{&ExternalSecretSecretStoreReferenceSpec{}, defaultExternalSecretSecretStoreReferenceSpec{}},
			want: &ExternalSecretSecretStoreReferenceSpec{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitializeExternalSecretSecretStoreReferenceSpec(tt.args.spec, tt.args.def); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeExternalSecretSecretStoreReferenceSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

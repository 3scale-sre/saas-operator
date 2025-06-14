package service

import (
	"strconv"
	"strings"

	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ELBServiceAnnotations returns annotations for services exposed through AWS Classic LoadBalancers
func ELBServiceAnnotations(cfg saasv1alpha1.ElasticLoadBalancerSpec, hostnames []string) map[string]string {
	annotations := map[string]string{
		"external-dns.alpha.kubernetes.io/hostname":                                      strings.Join(hostnames, ","),
		"service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled": strconv.FormatBool(*cfg.CrossZoneLoadBalancingEnabled),
		"service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled":       strconv.FormatBool(*cfg.ConnectionDrainingEnabled),
		"service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout":       strconv.Itoa(int(*cfg.ConnectionDrainingTimeout)),
		"service.beta.kubernetes.io/aws-load-balancer-healthcheck-healthy-threshold":     strconv.Itoa(int(*cfg.HealthcheckHealthyThreshold)),
		"service.beta.kubernetes.io/aws-load-balancer-healthcheck-unhealthy-threshold":   strconv.Itoa(int(*cfg.HealthcheckUnhealthyThreshold)),
		"service.beta.kubernetes.io/aws-load-balancer-healthcheck-interval":              strconv.Itoa(int(*cfg.HealthcheckInterval)),
		"service.beta.kubernetes.io/aws-load-balancer-healthcheck-timeout":               strconv.Itoa(int(*cfg.HealthcheckTimeout)),
	}

	if *cfg.ProxyProtocol {
		annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"] = "*"
	}

	return annotations
}

// NLBServiceAnnotations returns annotations for services exposed through AWS Network LoadBalancers
func NLBServiceAnnotations(cfg saasv1alpha1.NetworkLoadBalancerSpec, hostnames []string) map[string]string {
	annotations := map[string]string{
		"external-dns.alpha.kubernetes.io/hostname":                    strings.Join(hostnames, ","),
		"service.beta.kubernetes.io/aws-load-balancer-type":            "external",
		"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": "instance",
		"service.beta.kubernetes.io/aws-load-balancer-scheme":          "internet-facing",
	}
	if *cfg.ProxyProtocol {
		annotations["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"] = "*"
	}

	if len(cfg.EIPAllocations) != 0 {
		annotations["service.beta.kubernetes.io/aws-load-balancer-eip-allocations"] = strings.Join(cfg.EIPAllocations, ",")
	}

	if cfg.LoadBalancerName != nil {
		annotations["service.beta.kubernetes.io/aws-load-balancer-name"] = *cfg.LoadBalancerName
	}

	attributes := []string{}
	if *cfg.CrossZoneLoadBalancingEnabled {
		attributes = append(attributes, "load_balancing.cross_zone.enabled=true")
	} else {
		attributes = append(attributes, "load_balancing.cross_zone.enabled=false")
	}

	if *cfg.DeletionProtection {
		attributes = append(attributes, "deletion_protection.enabled=true")
	} else {
		attributes = append(attributes, "deletion_protection.enabled=false")
	}

	annotations["service.beta.kubernetes.io/aws-load-balancer-attributes"] = strings.Join(attributes, ",")

	return annotations
}

// TCPPort returns a TCP corev1.ServicePort
func TCPPort(name string, port int32, targetPort intstr.IntOrString) corev1.ServicePort {
	return corev1.ServicePort{
		Name:       name,
		Port:       port,
		TargetPort: targetPort,
		Protocol:   corev1.ProtocolTCP,
	}
}

// Ports returns a list of corev1.ServicePort
func Ports(ports ...corev1.ServicePort) []corev1.ServicePort {
	list := []corev1.ServicePort{}

	return append(list, ports...)
}

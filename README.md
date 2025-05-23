# 3scale SaaS Operator

![3scale-saas](docs/logos/3scale-saas-logo.svg)

[![test](https://github.com/3scale-sre/saas-operator/actions/workflows/test.yaml/badge.svg)](https://github.com/3scale-sre/saas-operator/actions/workflows/test.yaml)
[![release](https://badgen.net/github/release/3scale/saas-operator)](https://github.com/3scale-sre/saas-operator/releases)
[![license](https://badgen.net/github/license/3scale/saas-operator)](https://github.com/3scale-sre/saas-operator/LICENSE)

A Kubernetes Operator based on the Operator SDK to manage 3scale SaaS (hosted version) on **Kubernetes/OpenShift**.

3scale SaaS controllers supported:

* Apicast
* AutoSSL
* Backend
* CORSProxy
* EchoAPI
* MappingService
* System
* Zync

## Requirements

* [prometheus-operator](https://github.com/coreos/prometheus-operator) v0.17.0+
* [grafana-operator](https://github.com/integr8ly/grafana-operator) v3.0.0+
* [External-secrets-operator](https://github.com/external-secrets/external-secrets) v0.4.4+
* [marin3r](https://github.com/3scale/marin3r) v0.7.0+
* [aws-nlb-helper-operator](https://github.com/3scale/aws-nlb-helper-operator) v0.2.0+

## Documentation

* [Getting started](docs/getting-started.md)
* [Log configuration](docs/logging.md)
* [Development](docs/development.md)
* [Release](docs/release.md)

### Custom resources reference

* [Apicast Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-apicast)
* [AutoSSL Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-autossl)
* [Backend Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-backend)
* [CORSProxy Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-corsproxy)
* [EchoAPI Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-echoapi)
* [MappingService Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-mappingservice)
* [System Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-system)
* [Zync Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-zync)
* [Sentinel Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-sentinel)
* [TwemproxyConfig Custom Resource Reference](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-twemproxyconfig)
* [RedisShard Custom Resource Reference (testing purpose only)](docs/api-reference/reference.asciidoc#k8s-api-github-com-3scale-saas-operator-api-v1alpha1-redisshard)

## License

3scale SaaS Operator is under Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
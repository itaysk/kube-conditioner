# Notice - This is an experimentation project, use for education only

# kube-conditioner

kube-conditioner is a Kubernetes controller that allows you to define custom Kubernetes Pod conditions based on your own logic.  

## Quickstart

After installing kube-controller (see below) in your Kubernetes cluster, create a CRD like:

```yaml
apiVersion: conditioner.itaysk.com/v1alpha1
kind: PodCondition
metadata:
  name: conditiontest
spec:
  interval: 5000
  labelSelector:
    matchLabels:
      itaysk: test
  prometheusSource:
    serverUrl: "http://my.prom.server:9090"
    rule: "myimportantmetric<50"
```
(see ./config/samples for sample yaml files)

This will:
- Add a new Pod condition called `conditiontest`
- To all Pods with label `itaysk: test`
- Determine the status of the by querying `Prometheus`
  - at `http://my.prom.server:9090`
  - with the query `myimportantmetric<50`
  - expecting the query to succeed (result is non empty)
- Evaluate the condition every `5000` milliseconds

## Motivation

Pod conditions are standard part of the Pod status section. They are described as:

> Conditions represent the latest available observations of an object's state. They are an extension mechanism intended to be used when the details of an observation are not a priori known or would not apply to all instances of a given Kind
[source](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status)

Use cases:
- As a handy customizable sub-status on your pods that you can query manually or from scripts.
- To determine the Pod readiness state using [Readiness Gate](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-readiness-gate).
- For automatic remediation scenarios.

For more context and information, please see this blog post: [http://blog.itaysk.com/2019/03/27/kube-conditioner-custom-pod-conditions-for-everyone](http://blog.itaysk.com/2019/03/27/kube-conditioner-custom-pod-conditions-for-everyone)

## Datasources

The controller is pluggable by design in regards to how to determin the status of the condition. The current implementation supports running queries against Prometheus in a pretty basic way. This area needs work and additional datasources may be added.

The Prometheus datasource is documented under [./pkg/datasource/prometheus/Readme.md](./pkg/datasource/prometheus/Readme.md)

## Installing

There's an image in docker hub and generated deployment files for easy experimentation but the process is not yet automated to get the latest bits it's better to build from source and generate deployment files. 

### From distribution

```bash
kubectl apply -f config/crds
kubectl apply -f https://raw.githubusercontent.com/itaysk/kube-conditioner/master/kube-deploy.yaml
```

### From source

install dependencies:

- go
- dep
- kustomize
- kubebuilder
- docker

```bash
#clone
git clone https://github.com/itaysk/kube-conditioner.git
cd kube-conditioner
#build
make
IMG=mydocker/kube-builder:latest
make docker-build
#deploy
make docker-push
make deploy
```

## Contributing

PRs/issues are welcome!

Built using:
- go 1.11.5
- kube-builder 1.0.8
- Kubernetes 1.13.2 (minikube)


[![Go Report Card](https://goreportcard.com/badge/github.com/lightbend/akka-cluster-operator)](https://goreportcard.com/report/github.com/lightbend/akka-cluster-operator)

# akka-cluster-operator

Kubernetes operator for applications using Akka clustering. This is demo-ware, a strawman for
looking at the value proposition, and is not meant for real use.

This takes an AkkaCluster resource specification, which is the same as a regular Deployment specification,
then installs

* a ServiceAccount to allow the application to list its own pods. Note that this does _not_ change the default
serviceaccount in the namespace, and every AkkaCluster application has its own serviceaccount.

* a Role to be a pod-reader

* RoleBinding to connect the serviceaccount to the role

* Deployment per specification but with default ServiceAccount, pod selector, rolling update strategy, and
AKKA_CLUSTER_BOOTSTRAP_SERVICE_NAME environment settings.

Then on the Status side, the AkkaCluster status reflects the Akka Management endpoint, so one can use normal kubernetes
tools like `kubectl` or `oc` or the cluster UI to look at Akka Cluster leader, members, connection problems, etc.

## install the CRD

this needs to be done once per cluster

```bash
kubectl apply -f ./deploy/crds/app_v1alpha1_akkacluster_crd.yaml
```

## install the controller

in each namespace where akkacluster apps are desired

```bash
kubectl apply -f ./deploy
```

## demo application

https://github.com/dbrinegar/akka-java-cluster-openshift

## hacking

* install operator-sdk
* start minikube
* install the CRD
* route pod network to macbook so operator can query akka management endpoints `sudo route -n add 172.17.0.0/16 $(minikube ip)`

then loop on:
* `operator-sdk up local`

and a demo app


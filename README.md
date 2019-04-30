# akka-cluster-operator

Kubernetes operator for applications using Akka clustering

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

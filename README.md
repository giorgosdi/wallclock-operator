# Wallclock Operator for Kubernetes

This repository is just for up-skilling purposes with kubernetes operators and the Operator SDK in general.

## Summary

The Wallclock operator creates two custom resources:
- `Timezones`
- `Wallclocks`

The manifests look like the following:

**Timezones**
```YAML 
apiVersion: clock.giorgos.com/v1
kind: Timezones
metadata:
  name: example-timezones
spec:
  clocks: 
    - GMT
    - CET
    - PST8PDT
```
**Wallclocks**
```YAML
apiVersion: clock.giorgos.com/v1
kind: WallClock
metadata:
    name: example-timezones-pdt
spec:
    timezone: PDT
status:
    time: "15:00:00"
```
Whenever the operator finds out the there is one or more `Timezone` resources, it will check for its Spec to find out which `Clocks` exist.

It will then take the value of each clock (`GMT`, `CET` etc.) and create a `Wallclock` with the timezone and the actual time of the timezone based on the current UTC time.

## Timezones

The timezones that are used for this operator are based on timezone data on the OS itself. 

The `alpine` has been used to develop this operator. The image iteself has no timezone data by default, thats why you need to install them from your Dockerfile.

The dockerfile for this repo already does it for you using **apk**:

`apk add tzdata`

Once you install the timezone data in alpine, you can check them out under:

`/usr/share/zoneinfo/`

## Local development

For local developemnt you will need a couple of things to get you starts.

1) [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
2) [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
2) [Operator SDK](https://github.com/operator-framework/operator-sdk)

You can start minikube with this command : `minikube start`

Minikube will start a small local cluster that exists for testing purposes.

When your cluster is ready you should run the following commands to setup the cluster and the wallclock operator:

Setup Service Account:

`kubectl create -f deploy/service_account.yaml`

Setup RBAC

`kubectl create -f deploy/role.yaml`
`kubectl create -f deploy/role_binding.yaml`

Setup Wallclock and Timezones

`kubectl create -f deploy/crds/clock_v1_wallclock_crd.yaml`
`kubectl create -f deploy/crds/clock_v1_timezones_crd.yaml`

Create the operator

`kubectl create -f deploy/operator.yaml`



## Testing


You can then start creating `Timezones` resource like so:

`kubectl apply -f deploy/crds/clock_v1_timezones_cr.yaml`

and the operator will start creating `Wallclock` resources based on the specification you provided.

You can check out the `Wallclock` resources like this:

`kubectl get wallclocks`
# Inspektor Gadget on OpenShift

## Installation

Authenticate to your favorite cluster and apply the following command:

```
oc create -f openshift/deployment.yaml
```

## Changes against standard version

On OpenShift I try to avoid:
* adding stuff in the kube-system namespace so gadget-tracing namespace is created
* Dealing directly with host system directories (/run /sys ...). So sockets and ebpf maps are relocated to less sensitive directories
* Use fedora instead of alpine for the container to be more aligned to Red Hat way of doing things.

## How to build

```
export REGISTRY=<your registry>
make SUDO= DOCKER=podman CONTAINER_REPO=$REGISTRY/gadget-tracing/gadget
```

Then adapt the openshift/deployment.yaml in consequences to grab the image from your registry.


## TODO

* find why gadget daemon is not running on DaemonSet deployment (bug in entrypoint).
* integrate gadget with oc cli
* rebuild traceloop to use socket parameter to be able to use traceloop gadget

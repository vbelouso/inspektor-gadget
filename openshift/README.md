# Inspektor Gadget on OpenShift

## Installation

Authenticate to your favorite cluster and apply the following command:

```
oc create -f openshift/deployment.yaml
```

## Changes against standard version

On OpenShift I try to avoid:
- adding stuff in the kube-system namespace so gadget-tracing namespace is created
- Dealing directly with host system directories (/run /sys ...). So sockets and ebpf maps are relocated to less sensitive directories
- Use fedora instead of alpine for the container to be more aligned to Red Hat way of doing things.



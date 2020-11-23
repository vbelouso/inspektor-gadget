# Inspektor Gadget on OpenShift

## Installation

Authenticate to your favorite cluster and apply the following command:

```bash
oc apply -f https://raw.githubusercontent.com/clustership/inspektor-gadget/master/openshift/deployment.yaml
```

NOTE: There is a bug currently where the daemonset can fail with sigsev.
In that case, delete the daemonset and redeploy it again:

```bash
oc delete ds gadget -n gadget-tracing
oc apply -f https://raw.githubusercontent.com/clustership/inspektor-gadget/master/openshift/deployment.yaml
```

## Compile the client

It is required to have make, git and golang already installed. RHEL8:

```bash
sudo dnf install -y make git golang
```

Clone this repo:

```bash
git clone https://github.com/clustership/inspektor-gadget.git
```

Build the client:

```bash
make kubectl-gadget-linux-amd64
```

Copy the binary to the `/usr/local/bin/` directory so `oc` and `kubectl` can recognize it as a plugin:

```bash
chmod a+x ./kubectl-gadget-linux-amd64
sudo mv kubectl-gadget-linux-amd64 /usr/local/bin/kubectl-gadget
sudo restorecon /usr/local/bin/kubectl-gadget
```

## Usage

Deploy a demo application. Basically [this one](https://raw.githubusercontent.com/kinvolk/inspektor-gadget/master/docs/examples/app-set-priority.yaml) but in the `gadget-tracing` namespace:

```bash
cat <<EOF | oc apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: set-priority
  namespace: gadget-tracing
  labels:
    k8s-app: set-priority
spec:
  selector:
    matchLabels:
      name: set-priority
  template:
    metadata:
      labels:
        name: set-priority
    spec:
      containers:
      - name: set-priority
        image: busybox
        command: [ "sh", "-c", "while /bin/true ; do nice -n -20 echo ; sleep 5; done" ]
EOF
```

Then, get the pod name and the worker name where the pod is running. In this case, we will be focused on kni1-worker-0:

```bash
oc get po -o wide -n gadget-tracing | grep -i worker-0
gadget-6vp9b                    1/1     Running   0          21m   10.19.138.9    kni1-worker-0.example.com   <none>           <none>
set-priority-6844c9588d-5n29k   1/1     Running   0          21m   10.131.0.97    kni1-worker-0.example.com   <none>           <none>
```

In this case, we chose the set-priority pod running in kni1-worker-0, hence set-priority-6844c9588d-5n29k.

Run the kubectl-gadget capabilities command as:

```bash
kubectl-gadget capabilities --unique --verbose -p set-priority-6844c9588d-5n29k --node kni1-worker-0.example.com
```

or

```bash
oc gadget capabilities --verbose --unique -p set-priority-6844c9588d-5n29k --node kni1-worker-0.example.com
```

The output looks like:

```
Node numbers: 6 = kni1-worker-0.example.com
Running command: exec /opt/bcck8s/bcc-wrapper.sh --tracerid 20201123115948-c2009dae9253 --gadget /usr/share/bcc/tools/capable  --namespace gadget-tracing --podname set-priority-6844c9588d-5n29k  --  --unique -v
NODE TIME      UID    PID    COMM             CAP  NAME                 AUDIT 
[ 6] 11:59:50  1000670000 3342927 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:50  1000670000 3404871 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:50  1000670000 3404871 true             6    CAP_SETGID           1     
[ 6] 11:59:50  1000670000 3404871 true             7    CAP_SETUID           1     
[ 6] 11:59:50  1000670000 3404872 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:50  1000670000 3404872 nice             6    CAP_SETGID           1     
[ 6] 11:59:50  1000670000 3404872 nice             7    CAP_SETUID           1     
[ 6] 11:59:50  1000670000 3404872 nice             23   CAP_SYS_NICE         1     
[ 6] 11:59:50  1000670000 3404873 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:50  1000670000 3404873 sleep            6    CAP_SETGID           1     
[ 6] 11:59:50  1000670000 3404873 sleep            7    CAP_SETUID           1     
[ 6] 11:59:55  1000670000 3405107 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:55  1000670000 3405107 true             6    CAP_SETGID           1     
[ 6] 11:59:55  1000670000 3405107 true             7    CAP_SETUID           1     
[ 6] 11:59:55  1000670000 3405108 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:55  1000670000 3405108 nice             6    CAP_SETGID           1     
[ 6] 11:59:55  1000670000 3405108 nice             7    CAP_SETUID           1     
[ 6] 11:59:55  1000670000 3405108 nice             23   CAP_SYS_NICE         1     
[ 6] 11:59:56  1000670000 3405109 sh               21   CAP_SYS_ADMIN        0     
[ 6] 11:59:56  1000670000 3405109 sleep            6    CAP_SETGID           1     
[ 6] 11:59:56  1000670000 3405109 sleep            7    CAP_SETUID           1     
^C
Terminating...
Running command: exec /opt/bcck8s/bcc-wrapper.sh --tracerid 20201123115948-c2009dae9253 --stop
```

## Changes against standard version

On OpenShift I try to avoid:

* adding stuff in the kube-system namespace so gadget-tracing namespace is created
* Dealing directly with host system directories (/run /sys ...). So sockets and ebpf maps are relocated to less sensitive directories
* Use fedora instead of alpine for the container to be more aligned to Red Hat way of doing things.

## How to build

```bash
export REGISTRY=<your registry>
make SUDO= DOCKER=podman CONTAINER_REPO=$REGISTRY/gadget-tracing/gadget
```

Then adapt the openshift/deployment.yaml in consequences to grab the image from your registry.

## TODO

* find why gadget daemon is not running on DaemonSet deployment (bug in entrypoint).
* integrate gadget with oc cli
* rebuild traceloop to use socket parameter to be able to use traceloop gadget

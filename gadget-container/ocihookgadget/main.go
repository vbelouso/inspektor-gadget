package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "github.com/kinvolk/inspektor-gadget/pkg/gadgettracermanager/api"
	"github.com/kinvolk/inspektor-gadget/pkg/gadgettracermanager/containerutils"

	"github.com/containerd/nri/skel"
	types "github.com/containerd/nri/types/v1"
)

var (
	socketfile string
	hook       string
)

func init() {
	flag.StringVar(&socketfile, "socketfile", "/run/gadgettracermanager.socket", "Socket file")
	flag.StringVar(&hook, "hook", "", "OCI hook: prestart or poststop")
}

type igHook struct {
}

func (i *igHook) Type() string {
	return "ighook"
}

func (i *igHook) Invoke(ctx context.Context, r *types.Request) (*types.Result, error) {
	// TODO: Why the sandbox is leaked during delete (it's never deleted from the
	// gadgettracermanager)
	if !r.IsSandbox() {
		switch r.State {
		case types.Create:
			hook = "prestart"
			processContainer(r.ID, r.Pid)
		case types.Delete:
			hook = "poststop"
			processContainer(r.ID, r.Pid)
		}
	}

	result := r.NewResult("ighook")
	return result, nil
}

func main() {
	ctx := context.Background()
	if err := skel.Run(ctx, &igHook{}); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing ighook: %v", err)
		// don't return an error as it's a debug tool and we don't want to
		// create extra trouble if there is a failure.
		os.Exit(0)
	}
}

func processContainer(ociStateID string, ociStatePid int) error {
	// Validate state
	if ociStateID == "" || (ociStatePid == 0 && hook == "prestart") {
		return fmt.Errorf("invalid OCI state: %v %v", ociStateID, ociStatePid)
	}

	// Connect to the Gadget Tracer Manager
	var client pb.GadgetTracerManagerClient
	var ctx context.Context
	var cancel context.CancelFunc
	conn, err := grpc.Dial("unix://"+socketfile, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client = pb.NewGadgetTracerManagerClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Handle the poststop hook first
	if hook == "poststop" {
		_, err := client.RemoveContainer(ctx, &pb.ContainerDefinition{
			ContainerId: ociStateID,
		})
		if err != nil {
			return err
		}
		return nil
	}

	// TODO: NRI includes a cgroup path, can that be used?
	// Get cgroup paths
	cgroupPathV1, cgroupPathV2, err := containerutils.GetCgroupPaths(ociStatePid)
	if err != nil {
		return err
	}
	cgroupPathV2WithMountpoint, _ := containerutils.CgroupPathV2AddMountpoint(cgroupPathV2)

	// Get cgroup-v2 id
	cgroupId, _ := containerutils.GetCgroupID(cgroupPathV2WithMountpoint)

	// Get mount namespace ino
	mntns, err := containerutils.GetMntNs(ociStatePid)
	if err != nil {
		return err
	}

	//TODO: MountSources is missing and it's used to get the container name.
	// It seems that the NRI provides the container name directly under
	// .spec.annotations.io.kubernetes.cri.container-name

	_, err = client.AddContainer(ctx, &pb.ContainerDefinition{
		ContainerId:    ociStateID,
		CgroupPath:     cgroupPathV2WithMountpoint,
		CgroupId:       cgroupId,
		Mntns:          mntns,
		CgroupV1:       cgroupPathV1,
		CgroupV2:       cgroupPathV2,
		//MountSources:   mountSources,
	})
	if err != nil {
		return err
	}
	return nil
}

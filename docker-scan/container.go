package main

import (
	"context"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

type containerScanResults struct {
	ID              string
	Image           string
	Running         bool
	Privileged      bool
	PublishAllPorts bool
	HostMounts      []string
	CapAdd          []string
}

func scanContainer(ctx context.Context, client dockerclient.APIClient, c types.Container, wg *sync.WaitGroup, resultCh chan *containerScanResults, errCh chan error) {
	defer wg.Done()

	alert := false

	info, err := client.ContainerInspect(ctx, c.ID)
	if err != nil {
		errCh <- err
		return
	}
	res := &containerScanResults{
		ID:      c.ID[:12],
		Image:   c.Image,
		Running: info.State.Running,
	}
	// TODO: inspect container for risks
	if info.HostConfig.Privileged {
		alert = true
		res.Privileged = true
	}

	if len(info.HostConfig.CapAdd) > 0 {
		alert = true
		res.CapAdd = info.HostConfig.CapAdd
	}

	if len(info.HostConfig.Binds) > 0 {
		binds := []string{}
		for _, bind := range info.HostConfig.Binds {
			parts := strings.Split(bind, ":")
			if len(parts) != 2 {
				continue
			}

			source := parts[0]
			if strings.Index(source, "/") == 0 {
				binds = append(binds, source)
			}
		}
		res.HostMounts = binds
	}

	if v := info.HostConfig.PublishAllPorts; v {
		alert = true
		res.PublishAllPorts = v
	}

	if !alert {
		return
	}
	resultCh <- res
}

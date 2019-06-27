package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type imageScanConfig struct {
	Image      string
	BuildImage string
}

type imageScanSummary struct {
	Total      int `json:"total"`
	Low        int `json:"low"`
	Medium     int `json:"medium"`
	High       int `json:"high"`
	Negligible int `json:"negligible"`
	Sensitive  int `json:"sensitive"`
	Malware    int `json:"malware"`
}
type imageScanResults struct {
	Image   string
	Summary imageScanSummary `json:"vulnerability_summary"`
}

func scanImage(ctx context.Context, client dockerclient.APIClient, img, token string, noPull bool, wg *sync.WaitGroup, resultCh chan *imageScanResults, errCh chan error) {
	defer wg.Done()

	if !noPull {
		if _, err := client.ImagePull(ctx, img, types.ImagePullOptions{All: true}); err != nil {
			errCh <- err
			return
		}
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tmpl := template.New("dockerfile")
	t, err := tmpl.Parse(scanTemplate)
	if err != nil {
		errCh <- err
		return
	}

	// build the image to scan
	buildImage := fmt.Sprintf("scan-%d", time.Now().UnixNano())

	dockerfile := new(bytes.Buffer)
	if err := t.Execute(dockerfile, imageScanConfig{Image: img, BuildImage: buildImage}); err != nil {
		errCh <- err
		return
	}

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile.Bytes())),
	}
	if err = tw.WriteHeader(tarHeader); err != nil {
		errCh <- err
		return
	}
	if _, err = tw.Write(dockerfile.Bytes()); err != nil {
		errCh <- err
		return
	}
	dfr := bytes.NewReader(buf.Bytes())
	if _, err := client.ImageBuild(
		ctx,
		dfr,
		types.ImageBuildOptions{
			Dockerfile:     "Dockerfile",
			Tags:           []string{buildImage},
			PullParent:     true,
			Context:        dfr,
			NoCache:        true,
			SuppressOutput: true,
			BuildArgs: map[string]*string{
				"token": &token,
			},
		},
	); err != nil {
		errCh <- err
		return
	}

	// run the image to get the scan results
	resp, err := client.ContainerCreate(ctx, &container.Config{
		Image: buildImage,
		Tty:   false,
	}, nil, nil, "")
	if err != nil {
		errCh <- err
		return
	}
	defer func() {
		client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		client.ImageRemove(ctx, buildImage, types.ImageRemoveOptions{Force: true})
	}()

	if err := client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		errCh <- err
		return
	}

	logs := new(bytes.Buffer)

	out, err := client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Details:    false,
	})
	if err != nil {
		errCh <- err
		return
	}

	if _, err := stdcopy.StdCopy(logs, nil, out); err != nil {
		errCh <- err
		return
	}

	results := &imageScanResults{
		Image: img,
	}
	if err := json.NewDecoder(logs).Decode(results); err != nil {
		errCh <- err
		return
	}
	resultCh <- results
}

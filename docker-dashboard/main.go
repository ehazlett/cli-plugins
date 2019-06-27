package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/capnm/sysinfo"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	units "github.com/docker/go-units"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/cpu"
	"github.com/spf13/cobra"
)

const (
	infoText = `Time: %s
ID: %s
Name: %s
Docker Version: %s
OS: %s (%s)
Uptime: %s
Load: %v
`
	eventText = `ID:	%s
Date:	%s
Type:	%s
Action:	%s
Scope:	%s
`

	swarmText = `NodeID: %s
Address: %s
State: %s
Managers: %d
Nodes: %d
`
)

var (
	infoDisplay       = widgets.NewParagraph()
	cpuGauge          = widgets.NewGauge()
	memGauge          = widgets.NewGauge()
	resourceDisplay   = widgets.NewStackedBarChart()
	resourceData      = make([][]float64, 3)
	eventDisplay      = widgets.NewParagraph()
	containersDisplay = widgets.NewParagraph()
	swarmDisplay      = widgets.NewParagraph()
)

func getInfoText(info types.Info, si *sysinfo.SI) string {
	return fmt.Sprintf(infoText,
		time.Now().UTC().String(),
		info.ID,
		info.Name,
		info.ServerVersion,
		info.OperatingSystem,
		info.Architecture,
		units.HumanDuration(si.Uptime),
		si.Loads[0],
	)
}

func getEventText(e events.Message) string {
	t := time.Unix(e.Time, 0)
	return fmt.Sprintf(eventText, e.ID, t.UTC(), e.Type, e.Action, e.Scope)
}

func getContainersText(containers []types.Container) string {
	data := []string{}
	for _, c := range containers {
		data = append(data, fmt.Sprintf("ID: %s Image: %s", c.ID[:12], c.Image))
	}

	return strings.Join(data, "\n")
}

func getSwarmInfoText(info types.Info) string {
	if info.Swarm.NodeID == "" {
		return "Swarm not active"
	}
	s := info.Swarm
	return fmt.Sprintf(swarmText,
		s.NodeID,
		s.NodeAddr,
		s.LocalNodeState,
		s.Managers,
		s.Nodes,
	)
}

func update(ctx context.Context, cli dockerclient.APIClient) error {
	si := sysinfo.Get()
	info, err := cli.Info(ctx)
	if err != nil {
		return err
	}

	infoDisplay.Text = getInfoText(info, si)
	// TODO: update cpuGauge
	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	cpuGauge.Percent = int(cpuUsage[0])
	// TODO: update memGauge
	memGauge.Percent = int((si.TotalRam / si.FreeRam))

	volResp, err := cli.VolumeList(ctx, filters.NewArgs())
	if err != nil {
		return err
	}

	resourceData[0] = []float64{float64(info.Containers)}
	resourceData[1] = []float64{float64(info.Images)}
	resourceData[2] = []float64{float64(len(volResp.Volumes))}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		return err
	}
	containersDisplay.Text = getContainersText(containers)
	swarmDisplay.Text = getSwarmInfoText(info)
	return nil
}

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "dashboard",
			Short: "Docker Terminal Dashboard",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli := dockerCli.Client()
				ctx := context.Background()

				info, err := cli.Info(ctx)
				if err != nil {
					return err
				}

				si := sysinfo.Get()

				if err := ui.Init(); err != nil {
					return err

				}
				defer ui.Close()

				infoDisplay.Title = "Node Info"
				infoDisplay.Text = getInfoText(info, si)

				grid := ui.NewGrid()
				termWidth, termHeight := ui.TerminalDimensions()
				grid.SetRect(0, 0, termWidth, termHeight)
				// cpu
				cpuGauge.Title = "CPU"
				cpuGauge.BarColor = ui.ColorBlue
				// memory
				memGauge.Title = "Memory"
				memGauge.Percent = int((si.TotalRam / si.FreeRam))
				memGauge.BarColor = ui.ColorGreen
				// resources
				resourceDisplay.Title = "Resources"
				resourceDisplay.Data = resourceData
				resourceDisplay.Labels = []string{"Containers", "Images", "Volumes"}
				resourceDisplay.BarWidth = 25
				resourceDisplay.LabelStyles = []ui.Style{
					ui.NewStyle(ui.ColorWhite),
				}
				resourceDisplay.NumStyles = []ui.Style{
					ui.NewStyle(ui.ColorBlack),
				}
				resourceDisplay.BarColors = []ui.Color{
					ui.ColorCyan,
				}
				//events
				eventDisplay.Title = "Latest Event"
				// containers
				containersDisplay.Title = "Containers"
				// swarm
				swarmDisplay.Title = "Swarm Info"

				eventsCh, _ := cli.Events(ctx, types.EventsOptions{})
				go func() {
					for {
						select {
						case e := <-eventsCh:
							eventDisplay.Text = getEventText(e)
						}
					}
				}()

				grid.Set(
					ui.NewRow(1.0/2,
						ui.NewCol(1.0/4, infoDisplay),
						ui.NewCol(1.0/4, swarmDisplay),
						ui.NewCol(1.0/2,
							ui.NewRow(1.0/2, cpuGauge),
							ui.NewRow(1.0/2, memGauge),
						),
					),
					ui.NewRow(1.0/2,
						ui.NewCol(1.0/2, resourceDisplay),
						ui.NewCol(1.0/2,
							ui.NewRow(1.0/2,
								ui.NewCol(1.0, containersDisplay),
							),
							ui.NewRow(1.0/2,
								ui.NewCol(1.0, eventDisplay),
							)),
					),
				)

				ui.Render(grid)

				uiEvents := ui.PollEvents()
				ticker := time.NewTicker(time.Second).C
				for {
					select {
					case e := <-uiEvents:
						switch e.Type {
						case ui.KeyboardEvent:
							return nil
						}
					case <-ticker:
						if err := update(ctx, cli); err != nil {
							return err
						}
						ui.Render(grid)
					}

				}
				return nil
			},
		}

		return cmd
	},
		manager.Metadata{
			SchemaVersion: "0.1.0",
			Vendor:        "@ehazlett",
			Version:       "0.1.0",
		})
}

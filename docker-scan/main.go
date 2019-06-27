package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	noPull         bool
	imageScanToken string
)

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		images := &cobra.Command{
			Use:   "images",
			Short: "Scan images for vulnerabilities using Aquasec Microscanner",
			Args:  cli.RequiresMinArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cli := dockerCli.Client()
				ctx := context.Background()

				tw := tabwriter.NewWriter(os.Stdout, 1, 12, 1, ' ', 0)
				fmt.Fprintln(tw, "ID\tTOTAL\tLOW\tMEDIUM\tHIGH\tMALWARE\t")

				wg := &sync.WaitGroup{}

				errCh := make(chan error)
				go func() {
					for {
						select {
						case err := <-errCh:
							fmt.Fprintln(os.Stderr, err.Error())
						default:
						}
					}
				}()

				resultCh := make(chan *imageScanResults)
				go func() {
					for {
						select {
						case res := <-resultCh:
							fmt.Fprintf(tw, "%v\t%d\t%d\t%d\t%d\t%d\n",
								res.Image,
								res.Summary.Total,
								res.Summary.Low,
								res.Summary.Medium,
								res.Summary.High,
								res.Summary.Malware,
							)
						}
					}
				}()

				for _, img := range args {
					wg.Add(1)
					go scanImage(ctx, cli, img, imageScanToken, noPull, wg, resultCh, errCh)
				}

				wg.Wait()

				return tw.Flush()
			},
		}
		containers := &cobra.Command{
			Use:   "containers",
			Short: "Check containers for various risks",
			RunE: func(cmd *cobra.Command, args []string) error {
				logrus.SetLevel(logrus.DebugLevel)
				cli := dockerCli.Client()
				ctx := context.Background()

				tw := tabwriter.NewWriter(os.Stdout, 2, 12, 2, ' ', 0)
				fmt.Fprintln(tw, "ID\tIMAGE\tRUNNING\tPRIVILEGED\tPUBLISHEDALLPORTS\tHOSTMOUNTS\tCAPADD\t")

				wg := &sync.WaitGroup{}

				errCh := make(chan error)
				go func() {
					for {
						select {
						case err := <-errCh:
							fmt.Fprintln(os.Stderr, err.Error())
						default:
						}
					}
				}()

				resultCh := make(chan *containerScanResults)
				go func() {
					for {
						select {
						case res := <-resultCh:
							capAdd := strings.Join(res.CapAdd, ",")
							fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n",
								res.ID,
								res.Image,
								res.Running,
								res.Privileged,
								res.PublishAllPorts,
								len(res.HostMounts),
								capAdd,
							)
						}
					}
				}()

				containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
				if err != nil {
					return err
				}

				for _, c := range containers {
					wg.Add(1)
					go scanContainer(ctx, cli, c, wg, resultCh, errCh)
				}

				wg.Wait()

				return tw.Flush()
			},
		}
		cmd := &cobra.Command{
			Use:   "scan",
			Short: "Docker Security Scanning",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				if err := plugin.PersistentPreRunE(cmd, args); err != nil {
					return err
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Usage()
			},
		}

		imagesFlags := images.Flags()
		imagesFlags.StringVar(&imageScanToken, "token", "", "microscanner token")
		imagesFlags.BoolVar(&noPull, "no-pull", false, "disable image pulling")

		cmd.AddCommand(images)
		cmd.AddCommand(containers)
		return cmd
	},
		manager.Metadata{
			SchemaVersion: "0.1.0",
			Vendor:        "@ehazlett",
			Version:       "0.1.0",
		})
}

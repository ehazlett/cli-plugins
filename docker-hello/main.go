package main

import (
	"context"
	"os"
	"text/template"

	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	units "github.com/docker/go-units"
	"github.com/spf13/cobra"
)

const (
	infoTemplate = `ID		{{.ID}}
CPU		{{.NCPU}}
Memory		{{humansize .MemTotal}}
Containers	{{.Containers}}
Images		{{.Images}}
OS		{{.OperatingSystem}} ({{.Architecture}})
`
)

func humanSize(s int64) string {
	return units.HumanSize(float64(s))
}

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "hello",
			Short: "Example Docker CLI Plugin",
			RunE: func(cmd *cobra.Command, args []string) error {
				cli := dockerCli.Client()
				ctx := context.Background()

				info, err := cli.Info(ctx)
				if err != nil {
					return err
				}

				tmpl := template.New("info").Funcs(template.FuncMap{
					"humansize": humanSize,
				})
				t, err := tmpl.Parse(infoTemplate)
				if err != nil {
					return err
				}

				return t.Execute(os.Stdout, info)
			},
		}

		return cmd
	},
		manager.Metadata{
			SchemaVersion: "0.1.0",
			Vendor:        "@ehazlett",
			Version:       "dev",
		})
}

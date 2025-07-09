package main

import (
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/containers/libpod/pkg/util"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	killCommand cliconfig.KillValues

	killDescription = "The main process inside each container specified will be sent SIGKILL, or any signal specified with option --signal."
	_killCommand    = &cobra.Command{
		Use:   "kill [flags] CONTAINER [CONTAINER...]",
		Short: "Kill one or more running containers with a specific signal",
		Long:  killDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			killCommand.InputArgs = args
			killCommand.GlobalFlags = MainGlobalOpts
			killCommand.Remote = remoteclient
			return killCmd(&killCommand)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			return checkAllLatestAndCIDFile(cmd, args, false, false)
		},
		Example: `podman kill mywebserver
  podman kill 860a4b23
  podman kill --signal TERM ctrID`,
	}
)

func init() {
	killCommand.Command = _killCommand
	killCommand.SetHelpTemplate(HelpTemplate())
	killCommand.SetUsageTemplate(UsageTemplate())
	flags := killCommand.Flags()

	flags.BoolVarP(&killCommand.All, "all", "a", false, "Signal all running containers")
	flags.StringVarP(&killCommand.Signal, "signal", "s", "KILL", "Signal to send to the container")
	flags.BoolVarP(&killCommand.Latest, "latest", "l", false, "Act on the latest container podman is aware of")

	markFlagHiddenForRemoteClient("latest", flags)
}

// killCmd kills one or more containers with a signal
func killCmd(c *cliconfig.KillValues) error {
	if c.Bool("trace") {
		span, _ := opentracing.StartSpanFromContext(Ctx, "killCmd")
		defer span.Finish()
	}

	// Check if the signalString provided by the user is valid
	// Invalid signals will return err
	killSignal, err := util.ParseSignal(c.Signal)
	if err != nil {
		return err
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "could not get runtime")
	}
	defer runtime.DeferredShutdown(false)

	ok, failures, err := runtime.KillContainers(getContext(), c, killSignal)
	if err != nil {
		return err
	}
	return printCmdResults(ok, failures)
}

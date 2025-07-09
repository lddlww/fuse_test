package main

import (
	"fmt"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	podStartCommand     cliconfig.PodStartValues
	podStartDescription = `The pod name or ID can be used.

  All containers defined in the pod will be started.`
	_podStartCommand = &cobra.Command{
		Use:   "start [flags] POD [POD...]",
		Short: "Start one or more pods",
		Long:  podStartDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			podStartCommand.InputArgs = args
			podStartCommand.GlobalFlags = MainGlobalOpts
			podStartCommand.Remote = remoteclient
			return podStartCmd(&podStartCommand)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			return checkAllLatestAndCIDFile(cmd, args, false, false)
		},
		Example: `podman pod start podID
  podman pod start --latest
  podman pod start --all`,
	}
)

func init() {
	podStartCommand.Command = _podStartCommand
	podStartCommand.SetHelpTemplate(HelpTemplate())
	podStartCommand.SetUsageTemplate(UsageTemplate())
	flags := podStartCommand.Flags()
	flags.BoolVarP(&podStartCommand.All, "all", "a", false, "Start all pods")
	flags.BoolVarP(&podStartCommand.Latest, "latest", "l", false, "Start the latest pod podman is aware of")
	markFlagHiddenForRemoteClient("latest", flags)
}

func podStartCmd(c *cliconfig.PodStartValues) error {
	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "could not get runtime")
	}
	defer runtime.DeferredShutdown(false)

	podStartIDs, podStartErrors := runtime.StartPods(getContext(), c)
	for _, p := range podStartIDs {
		fmt.Println(p)
	}
	if len(podStartErrors) == 0 {
		return nil
	}
	// Grab the last error
	lastError := podStartErrors[len(podStartErrors)-1]
	// Remove the last error from the error slice
	podStartErrors = podStartErrors[:len(podStartErrors)-1]

	for _, err := range podStartErrors {
		logrus.Errorf("%q", err)
	}
	return lastError
}

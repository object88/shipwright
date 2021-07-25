package run

import (
	"context"

	"github.com/object88/shipwright/internal/cmd/common"
	"github.com/object88/shipwright/pkg/http"
	httpcliflags "github.com/object88/shipwright/pkg/http/cliflags"
	"github.com/object88/shipwright/pkg/http/probes"
	"github.com/object88/shipwright/pkg/http/router"
	v1 "github.com/object88/shipwright/pkg/http/router/v1"
	k8scliflags "github.com/object88/shipwright/pkg/k8s/cliflags"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	httpFlagMgr *httpcliflags.FlagManager
	k8sFlagMgr  *k8scliflags.FlagManager

	probe *probes.Probe
}

// CreateCommand returns the `run` Command
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use:   "run",
			Short: "run",
			Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs:  ca,
		httpFlagMgr: httpcliflags.New(),
		k8sFlagMgr:  k8scliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	c.probe = probes.New()
	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.probe, c.startHTTPServer, c.startControllerManager, c.startInformerManager)
}

func (c *command) startHTTPServer(ctx context.Context, r probes.Reporter) error {
	rts, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(c.probe, v1.Defaults(c.Log)))
	if err != nil {
		return err
	}

	cf, err := c.httpFlagMgr.HttpsCertFile()
	if err != nil {
		return err
	}
	kf, err := c.httpFlagMgr.HttpsKeyFile()
	if err != nil {
		return err
	}

	h := http.New(c.Log, rts, c.httpFlagMgr.HttpPort())
	if p := c.httpFlagMgr.HttpsPort(); p != 0 {
		if err = h.ConfigureTLS(p, cf, kf); err != nil {
			return err
		}
	}

	c.Log.Info("starting http")
	defer c.Log.Info("http complete")

	h.Serve(ctx, r)
	return nil
}
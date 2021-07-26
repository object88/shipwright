package run

import (
	"context"

	"github.com/object88/shipwright/internal/cmd/common"
	"github.com/object88/shipwright/internal/http"
	httpcliflags "github.com/object88/shipwright/internal/http/cliflags"
	"github.com/object88/shipwright/internal/http/probes"
	"github.com/object88/shipwright/internal/http/router"
	k8scliflags "github.com/object88/shipwright/internal/k8s/cliflags"
	webhookcliflags "github.com/object88/shipwright/internal/webhook/cliflags"
	"github.com/object88/shipwright/internal/webhook/routes"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	httpFlagMgr    *httpcliflags.FlagManager
	k8sFlagMgr     *k8scliflags.FlagManager
	webhookFlagMgr *webhookcliflags.FlagManager

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
		CommonArgs:     ca,
		httpFlagMgr:    httpcliflags.New(),
		k8sFlagMgr:     k8scliflags.New(),
		webhookFlagMgr: webhookcliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)
	c.webhookFlagMgr.ConfigureSecret(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	c.probe = probes.New()
	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.probe, c.startHTTPServer)
}

func (c *command) startHTTPServer(ctx context.Context, r probes.Reporter) error {
	rts, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(c.probe, routes.Defaults(c.Log, c.webhookFlagMgr.Secret())))
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

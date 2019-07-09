package setup

import (
	"os"
	"path/filepath"

	cmdutil "github.com/gramLabs/redsky/pkg/redskyctl/util"
	"github.com/spf13/cobra"
)

const (
	resetLong    = `The reset command will uninstall the Red Sky manifests.`
	resetExample = ``
)

func NewResetCommand(f cmdutil.Factory, ioStreams cmdutil.IOStreams) *cobra.Command {
	o := NewSetupOptions(ioStreams)

	cmd := &cobra.Command{
		Use:     "reset",
		Short:   "Uninstall Red Sky from a cluster",
		Long:    resetLong,
		Example: resetExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Run())
		},
	}

	o.AddFlags(cmd)

	return cmd
}

func (o *SetupOptions) resetCluster() error {
	bootstrapConfig, err := NewBootstrapResetConfig(o)
	if err != nil {
		return err
	}

	// A bootstrap dry run just means serialize the bootstrap config
	if o.Bootstrap && o.DryRun {
		return bootstrapConfig.Marshal(o.Out)
	}

	// Remove bootstrap objects and return if that was all we are doing
	deleteFromCluster(bootstrapConfig)
	if o.Bootstrap {
		return nil
	}

	// Ensure a partial bootstrap is cleaned up properly
	defer deleteFromCluster(bootstrapConfig)

	// Create the bootstrap config to initiate the uninstall job
	podWatch, err := createInCluster(bootstrapConfig)
	if podWatch != nil {
		defer podWatch.Stop()
	}
	if err != nil {
		return err
	}

	// Wait for the job to finish; ignore errors as we are having the namespace pulled out from under us
	_ = waitForJob(o.ClientSet.CoreV1().Pods(o.namespace), podWatch, nil, nil)

	return nil

}

func (o *SetupOptions) resetKustomize() error {
	// TODO Walk back through the array to clean up empty directories
	p := filepath.Join(kustomizePluginDir()...)
	return os.RemoveAll(p)
}

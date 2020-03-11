/*
Copyright 2019 GramLabs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reset

import (
	"context"
	"fmt"
	"io"

	"github.com/redskyops/redskyops-controller/internal/config"
	"github.com/redskyops/redskyops-controller/redskyctl/internal/commander"
	"github.com/redskyops/redskyops-controller/redskyctl/internal/commands/grant_permissions"
	"github.com/redskyops/redskyops-controller/redskyctl/internal/commands/initialize"
	"github.com/spf13/cobra"
)

// Options is the configuration for suggesting assignments
type Options struct {
	// Config is the Red Sky Configuration used to generate the controller manifests for reset
	Config *config.RedSkyConfig
	// IOStreams are used to access the standard process streams
	commander.IOStreams
}

func NewCommand(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Uninstall from a cluster",
		Long:  "Uninstall Red Sky Ops from a cluster",

		PreRun: commander.StreamsPreRun(&o.IOStreams),
		RunE:   commander.WithContextE(o.reset),
	}

	commander.ExitOnError(cmd)
	return cmd
}

func (o *Options) reset(ctx context.Context) error {
	// Delete the CRDs first to avoid issues with the controller being deleted before it can remove the finalizers
	deleteCRD, err := o.Config.Kubectl(ctx, "delete", "--ignore-not-found", "crd", "trials.redskyops.dev", "experiments.redskyops.dev")
	if err != nil {
		return err
	}
	deleteCRD.Stdout = o.Out
	deleteCRD.Stderr = o.ErrOut
	if err := deleteCRD.Run(); err != nil {
		return err
	}

	// Fork `kubectl delete` and get a pipe to write manifests to
	kubectlDelete, err := o.Config.Kubectl(ctx, "delete", "--ignore-not-found", "-f", "-")
	if err != nil {
		return err
	}
	kubectlDelete.Stdout = o.Out
	kubectlDelete.Stderr = o.ErrOut
	w, err := kubectlDelete.StdinPipe()
	if err != nil {
		return err
	}
	if err := kubectlDelete.Start(); err != nil {
		return err
	}

	// Generate all of the manifests
	// TODO How should we synchronize this? How should we check the close error?
	errChan := make(chan error)
	go func() {
		err := o.generateManifests(w)
		if err != nil {
			errChan <- err
		}

		// Close the stream (tells kubectl there are no more resources to apply) and the error channel (so we can wait on kubectl)
		_ = w.Close()
		close(errChan)
	}()
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Wait for everything to be deleted
	return kubectlDelete.Wait()
}

// generateManifests writes all of the initialization manifests to the supplied writer
func (o *Options) generateManifests(out io.Writer) error {
	if err := o.generateInstall(out); err != nil {
		return err
	}
	if err := o.generateBootstrapRole(out); err != nil {
		return err
	}
	return nil
}

func (o *Options) generateInstall(out io.Writer) error {
	opts := &initialize.GeneratorOptions{
		Config: o.Config,
	}
	cmd := initialize.NewGeneratorCommand(opts)
	return o.executeCommand(cmd, out)
}

func (o *Options) generateBootstrapRole(out io.Writer) error {
	opts := &grant_permissions.GeneratorOptions{
		Config: o.Config,
	}
	cmd := grant_permissions.NewGeneratorCommand(opts)
	return o.executeCommand(cmd, out)
}

// executeCommand runs the supplied command, send output to the writer
func (o *Options) executeCommand(cmd *cobra.Command, out io.Writer) error {
	// Prepare the command and execute it
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(o.ErrOut)
	if err := cmd.Execute(); err != nil {
		return err
	}

	// Since we are dumping the output of multiple generators into a single stream, insert a YAML document separator
	_, _ = fmt.Fprintln(out, "---")
	return nil
}

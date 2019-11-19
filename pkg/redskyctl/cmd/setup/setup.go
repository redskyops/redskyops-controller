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

package setup

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/redskyops/k8s-experiment/pkg/redskyctl/cmd/generate"
	cmdutil "github.com/redskyops/k8s-experiment/pkg/redskyctl/util"
)

func install(kubectl *cmdutil.Kubectl, namespace string, env io.Reader, cmd *exec.Cmd) error {
	buffer := &bytes.Buffer{}

	// Create generate install options
	installOpts := generate.NewGenerateInstallOptions(cmdutil.IOStreams{Out: buffer, ErrOut: cmd.Stderr})
	installOpts.Kubectl = kubectl
	installOpts.Env = env
	if namespace != "" {
		installOpts.Namespace = namespace
	}

	// Populate the buffer
	if err := installOpts.Run(); err != nil {
		return err
	}

	// Run the command
	cmd.Stdin = buffer
	return cmd.Run()
}

func bootstrapRole(kubectl *cmdutil.Kubectl, cmd *exec.Cmd) error {
	buffer := &bytes.Buffer{}

	// Create generate RBAC options
	rbacOpts := generate.NewGenerateRBACOptions(cmdutil.IOStreams{Out: buffer, ErrOut: cmd.Stderr})
	rbacOpts.Bootstrap = true
	if err := rbacOpts.Complete(); err != nil {
		return err
	}

	// Populate the buffer
	if err := rbacOpts.Run(); err != nil {
		return err
	}

	// Run the command
	cmd.Stdin = buffer
	return cmd.Run()
}
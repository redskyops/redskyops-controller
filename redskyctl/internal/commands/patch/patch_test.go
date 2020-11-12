/*
Copyright 2020 GramLabs, Inc.

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

package patch_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/redskyops/redskyops-controller/redskyctl/internal/commander"
	"github.com/redskyops/redskyops-controller/redskyctl/internal/commands/patch"
	"github.com/redskyops/redskyops-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	_, expBytes, expFile := createTempExperimentFile(t)
	defer os.Remove(expFile.Name())

	manifestFile := createTempManifests(t)
	defer os.Remove(manifestFile.Name())

	testCases := []struct {
		desc  string
		args  []string
		stdin io.Reader
	}{
		{
			desc: "exp file manifest file",
			args: []string{
				"--file", expFile.Name(),
				"--file", manifestFile.Name(),
				"--trialname", "sampleExperiment-1234",
			},
		},
		{
			desc: "exp stdin manifest file",
			args: []string{
				"--file", "-",
				"--file", manifestFile.Name(),
				"--trialname", "sampleExperiment-1234",
			},
			stdin: bytes.NewReader(expBytes),
		},
		{
			desc: "exp file manifest stdin",
			args: []string{
				"--file", expFile.Name(),
				"--file", "-",
				"--trialname", "sampleExperiment-1234",
			},
			stdin: bytes.NewReader(pgDeployment),
		},
		{
			desc: "exp stdin manifest stdin",
			args: []string{
				"--file", "-",
				"--trialname", "sampleExperiment-1234",
			},
			stdin: bytes.NewReader(append(expBytes, pgDeployment...)),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%q", tc.desc), func(t *testing.T) {
			cfg := &config.RedSkyConfig{}
			opts := &patch.Options{Config: cfg}
			opts.ExperimentsAPI = &fakeRedSkyServer{}
			cmd := patch.NewCommand(opts)
			commander.ConfigGlobals(cfg, cmd)

			// setup output
			var b bytes.Buffer
			cmd.SetOut(&b)

			// setup input
			if tc.stdin != nil {
				cmd.SetIn(tc.stdin)
			}

			// set command args
			if len(tc.args) > 0 {
				cmd.SetArgs(tc.args)
			}

			err := cmd.Execute()
			require.NoError(t, err)

			cpu := wannabeTrial.TrialAssignments.Assignments[0]
			mem := wannabeTrial.TrialAssignments.Assignments[1]
			assert.Contains(t, b.String(), fmt.Sprintf("%s: %sm", cpu.ParameterName, cpu.Value.String()))
			assert.Contains(t, b.String(), fmt.Sprintf("%s: %sMi", mem.ParameterName, mem.Value.String()))
		})
	}
}

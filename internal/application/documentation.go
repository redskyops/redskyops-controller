/*
Copyright 2021 GramLabs, Inc.

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

package application

import (
	"os"
	"path/filepath"
	"strings"

	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// headComments is the mapping of field names to desired comments.
var headComments = map[string]string{

	"resources": `
Resources define where you application's Kubernetes resources come from. These
can be URL-like values such as file paths, HTTP URLs, or Git repository URLs.
They can also be more complex definitions such references to in-cluster objects
or Helm charts.
# Reference: https://docs.stormforge.io/reference/application/v1alpha1/#application
`,

	"parameters": `
Parameters are what our machine learning tunes in order to optimize your
application settings. You can optionally filter where to discover parameters by
using the selector (the default selector is "").
Reference: https://docs.stormforge.io/reference/application/v1alpha1/#parameters
`,

	"scenarios": `
Scenarios determine which load test will be used to put your application under
load during the experiment. You can create one by visiting https://app.stormforger.com/
Reference: https://docs.stormforge.io/reference/application/v1alpha1/#scenario
`,

	"objectives": `
Objectives are used to define what you want to optimize for. It's best to
optimize for metrics with inherent trade-offs such as cost and performance.
Objectives correspond to metrics observed over the course of a trial,
for example: "p95-latency".
Reference: https://docs.stormforge.io/reference/application/v1alpha1/#objective
`,
}

// DocumentationFilter looks for Application instances and attempts to annotate them
// with comments that might help people finish writing their app.yaml.
type DocumentationFilter struct {
	// Flag to completely disable documentation.
	Disabled bool
}

// Filter applies documentation to any applications in the supplied node set.
func (f *DocumentationFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	if f.Disabled {
		return nodes, nil
	}

	for _, n := range nodes {
		meta, err := n.GetMeta()
		if err != nil {
			continue
		}

		if meta.Kind != "Application" || meta.APIVersion != redskyappsv1alpha1.GroupVersion.String() {
			continue
		}

		if err := f.annotateApplication(n); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

// annotateApplication documents a single application.
func (f *DocumentationFilter) annotateApplication(app *yaml.RNode) error {
	n := app.YNode()

	// Reminder about how the file was generated
	if len(os.Args) >= 3 && filepath.Base(os.Args[0]) == "redskyctl" {
		n.HeadComment = strings.Join(os.Args, " ")
	}

	// Required is a map of field name to required preceding field name.
	// This is used to ensure even empty (and otherwise omitted) fields can be
	// included for documentation purposes.
	required := map[string]string{
		"resources":  "",
		"parameters": "resources",
		"scenarios":  "parameters",
		"objectives": "scenarios",
		"":           "objectives",
	}

	// Each key and value are elements in the content list, iterate over even indices
	var content []*yaml.Node
	for i := 0; i < len(n.Content); i = yaml.IncrementFieldIndex(i) {
		n.Content[i].HeadComment = headComments[n.Content[i].Value]
		content = append(content, missingRequiredContent(n.Content[i].Value, required)...)
		content = append(content, n.Content[i], n.Content[i+1])
	}

	// Make sure all the required content has been produced
	n.Content = append(content, missingRequiredContent("", required)...)

	return nil
}

// missingRequiredContent adds top level fields which may have been empty and therefore
// emitted during the initial Go to YAML encoding process. The supplied map is used
// to track which fields have already been encountered and which fields must precede
// the current field (identified by "key"). The resulting list of nodes are suitable
// for inclusion in the content of a mapping node.
func missingRequiredContent(key string, required map[string]string) []*yaml.Node {
	// As soon as we encounter a key, remove it so it does not get double added
	for k, v := range required {
		if v == key {
			delete(required, k)
		}
	}

	// Check if there is any required content for the key
	req := required[key]
	if req == "" {
		return nil
	}

	// Recursively include missing content first
	var result []*yaml.Node
	result = append(result, missingRequiredContent(req, required)...)

	// Add field name with the appropriate head comment
	result = append(result, &yaml.Node{
		Kind:        yaml.ScalarNode,
		Tag:         yaml.NodeTagString,
		Value:       req,
		HeadComment: headComments[req],
	})

	// Add an empty value
	switch req {
	case "parameters":
		result = append(result, &yaml.Node{
			Kind:  yaml.MappingNode,
			Style: yaml.FoldedStyle,
			Tag:   yaml.NodeTagMap,
		})

	default:
		result = append(result, &yaml.Node{
			Kind:  yaml.SequenceNode,
			Style: yaml.FoldedStyle,
			Tag:   yaml.NodeTagSeq,
		})
	}

	return result
}

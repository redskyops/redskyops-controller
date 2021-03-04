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

package generation

import (
	"bytes"
	"regexp"
	"strings"

	redskyv1beta1 "github.com/thestormforge/optimize-controller/api/v1beta1"
	"github.com/thestormforge/optimize-controller/internal/scan"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ParameterNamer is used to compute the name of an optimization parameter.
type ParameterNamer func(meta yaml.ResourceMeta, path []string, name string) string

// ExperimentSource allows selectors to modify the experiment directly. Note that
// for adding parameters, patches, or metrics, the appropriate source should be
// used instead.
type ExperimentSource interface {
	Update(exp *redskyv1beta1.Experiment) error
}

// ParameterSource allows selectors to add parameters to an experiment. In
// general PatchSources should also be ParameterSources to ensure the parameters
// used in the generated patches are configured on the experiment.
type ParameterSource interface {
	Parameters(name ParameterNamer) ([]redskyv1beta1.Parameter, error)
}

// PatchSource allows selectors to contribute changes to the patch of a
// particular resource. In general, ParameterSource should also be implemented
// to add any parameters referenced by the generated patches.
type PatchSource interface {
	TargetRef() *corev1.ObjectReference
	Patch(name ParameterNamer) (yaml.Filter, error)
}

// MetricSource allows selectors to contribute metrics to an experiment.
type MetricSource interface {
	Metrics() ([]redskyv1beta1.Metric, error)
}

// Transformer is used to convert all of the output from the selectors, only selector output
// matching the "*Source" interfaces are supported.
type Transformer struct {
	// The default name to use for the experiment (any ExperimentSource could theoretically override it).
	DefaultExperimentName string
	// Perform a merge on the result stream to account for duplicate resources.
	MergeGenerated bool
	// Flag indicating the all of the resources that were scanned should also be included in the output.
	IncludeApplicationResources bool
}

var _ scan.Transformer = &Transformer{}

func (t *Transformer) Filter(nodes []*yaml.RNode, selected []interface{}) ([]*yaml.RNode, error) {
	var result []*yaml.RNode

	// Parameter names need to be computed based on what resources were selected by the scan
	name := parameterNamer(selected)

	exp := redskyv1beta1.Experiment{}
	patches := make(map[corev1.ObjectReference][]yaml.Filter)
	for _, sel := range selected {
		// DO NOT use a type switch, there may be multiple implementations

		if e, ok := sel.(ExperimentSource); ok {
			if err := e.Update(&exp); err != nil {
				return nil, err
			}
		}

		if ps, ok := sel.(ParameterSource); ok {
			params, err := ps.Parameters(name)
			if err != nil {
				return nil, err
			}
			exp.Spec.Parameters = append(exp.Spec.Parameters, params...)
		}

		if ps, ok := sel.(PatchSource); ok {
			ref := ps.TargetRef()
			fs := patches[*ref]
			f, err := ps.Patch(name)
			if err != nil {
				return nil, err
			}
			patches[*ref] = append(fs, f)
		}

		if ms, ok := sel.(MetricSource); ok {
			metrics, err := ms.Metrics()
			if err != nil {
				return nil, err
			}
			exp.Spec.Metrics = append(exp.Spec.Metrics, metrics...)
		}

		// Also allow direct additions to the resource stream
		if r, ok := sel.(kio.Reader); ok {
			nodes, err := r.Read()
			if err != nil {
				return nil, err
			}
			result = append(result, nodes...)
		}
	}

	// Render patches into the experiment
	if err := t.renderPatches(patches, &exp); err != nil {
		return nil, err
	}

	// Set the default name on the experiment
	if exp.Name == "" {
		exp.Name = t.DefaultExperimentName
	}

	// Serialize the experiment as a YAML node
	if expNode, err := (scan.ObjectSlice{&exp}).Read(); err != nil {
		return nil, err
	} else {
		result = append(expNode, result...) // Put the experiment at the front
	}

	// Only merge the generate resources if it is configured. This may be necessary
	// to deal with duplication in the result stream, for example, if multiple
	// scenarios are present on the application at the time of generation.
	if t.MergeGenerated {
		if merged, err := (&filters.MergeFilter{}).Filter(result); err == nil {
			result = merged
		}
	}

	// TODO We should annotate everything up to this point as having been generated...

	// If requested, append the actual application resources to the output
	if t.IncludeApplicationResources {
		appResources, err := kio.FilterAll(yaml.SetAnnotation(filters.FmtAnnotation, filters.FmtStrategyNone)).Filter(nodes)
		if err != nil {
			return nil, err
		}
		result = append(result, appResources...)
	}

	return result, nil
}

// renderPatches converts accumulated patch contributes (in the form of yaml.Filter instances) into
// actual patch templates on an experiment.
func (t *Transformer) renderPatches(patches map[corev1.ObjectReference][]yaml.Filter, exp *redskyv1beta1.Experiment) error {
	for ref, fs := range patches {
		// Start with an empty node
		patch := yaml.NewRNode(&yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{{Kind: yaml.MappingNode}},
		})

		// Add of the patch contributes by executing the filters
		if err := patch.PipeE(fs...); err != nil {
			return err
		}

		// Render the result as YAML
		var buf bytes.Buffer
		if err := yaml.NewEncoder(&buf).Encode(patch.Document()); err != nil {
			return err
		}

		// Since the patch template doesn't need to be valid YAML we can cleanup tagged integers
		data := regexp.MustCompile(`!!int '(.*)'`).ReplaceAll(buf.Bytes(), []byte("$1"))

		// Add the actual patch to the experiment
		exp.Spec.Patches = append(exp.Spec.Patches, redskyv1beta1.PatchTemplate{
			Patch:     string(data),
			TargetRef: &ref,
		})
	}

	return nil
}

// pnode is the location and current state of something to parameterize in an application resource.
type pnode struct {
	meta      yaml.ResourceMeta // TODO Do we need the labels? Can this just be ResourceIdentifier?
	fieldPath []string
	value     *yaml.Node
}

// TargetRef returns the reference to the resource this parameter node belongs to.
func (p *pnode) TargetRef() *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: p.meta.APIVersion,
		Kind:       p.meta.Kind,
		Name:       p.meta.Name,
		Namespace:  p.meta.Namespace,
	}
}

// parameterNamer returns a name generation function for parameters based on scan results.
func parameterNamer(selected []interface{}) ParameterNamer {
	// Index the object references by kind and name
	type targeted interface {
		TargetRef() *corev1.ObjectReference
	}
	needsPath := make(map[string]map[string]int)
	for _, sel := range selected {
		t, ok := sel.(targeted)
		if !ok {
			continue
		}
		targetRef := t.TargetRef()
		if ns := needsPath[targetRef.Kind]; ns == nil {
			needsPath[targetRef.Kind] = make(map[string]int)
		}
		needsPath[targetRef.Kind][targetRef.Name]++
	}

	// Determine which prefixes we need
	needsKind := len(needsPath) > 1
	needsName := false
	for _, v := range needsPath {
		needsName = needsName || len(v) > 1
	}

	return func(meta yaml.ResourceMeta, path []string, name string) string {
		var parts []string
		if needsKind {
			parts = append(parts, strings.ToLower(meta.Kind))
		}
		if needsName {
			parts = append(parts, strings.Split(meta.Name, "-")...)
		}
		if needsPath[meta.Kind][meta.Name] > 1 {
			for _, p := range path {
				if yaml.IsListIndex(p) {
					if _, value, _ := yaml.SplitIndexNameValue(p); value != "" {
						parts = append(parts, value) // TODO Split on "-" like we do for names?
					}
				}
			}
		}
		return strings.Join(append(parts, name), "_")
	}
}

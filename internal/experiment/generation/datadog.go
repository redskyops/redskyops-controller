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

package generation

import (
	"net/url"

	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	redskyv1beta1 "github.com/thestormforge/optimize-controller/api/v1beta1"
)

type DatadogMetricsSource struct {
	Goal *redskyappsv1alpha1.Goal
}

var _ MetricSource = &DatadogMetricsSource{}

func (s *DatadogMetricsSource) Metrics() ([]redskyv1beta1.Metric, error) {
	var result []redskyv1beta1.Metric
	if s.Goal == nil || s.Goal.Implemented {
		return result, nil
	}

	m := newGoalMetric(s.Goal, s.Goal.Datadog.Query)
	m.Type = redskyv1beta1.MetricDatadog
	m.Minimize = !s.Goal.Datadog.Maximize
	if s.Goal.Datadog.Aggregator != "" {
		m.URL = "?" + url.Values{"aggregator": []string{s.Goal.Datadog.Aggregator}}.Encode()
	}
	result = append(result, m)

	return result, nil
}
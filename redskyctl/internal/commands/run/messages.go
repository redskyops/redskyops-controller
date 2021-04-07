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

package run

import (
	"time"

	"github.com/thestormforge/konjure/pkg/konjure"
	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	"github.com/thestormforge/optimize-controller/internal/version"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const unknownVersion = "unknown"

type versionMsg struct {
	Build   version.Info
	Forge   string
	Kubectl struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}
	Controller version.Info
}

type authorizationMsg int

const (
	azValid   authorizationMsg = 1
	azInvalid authorizationMsg = 2
	azIgnored authorizationMsg = 3
)

type stormForgerTestCaseMsg []string

type kubernetesNamespaceMsg []string

type resourceMsg []konjure.Resource

type scenarioMsg []redskyappsv1alpha1.Scenario

type parameterMsg []redskyappsv1alpha1.Parameter

type objectiveMsg []redskyappsv1alpha1.Objective

type ingressMsg redskyappsv1alpha1.Ingress

type experimentMsg []*yaml.RNode

type trialsMsg []*yaml.RNode

type refreshStatusMsg time.Time

type experimentStatusMsg int

const (
	expPreCreate experimentStatusMsg = 1
	expConfirmed experimentStatusMsg = 2
	expCreated   experimentStatusMsg = 3
	expCompleted experimentStatusMsg = 4
	expFailed    experimentStatusMsg = 5
)
// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istiooperator //nolint:testpackage

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
)

func TestConversion(t *testing.T) {
	testCases := []struct {
		name                 string
		inputConfig          string
		expectedManifestName testutil.ObjectMetadata
		expected             string
		jsonPath             string
	}{
		{
			name:        "default profile",
			inputConfig: `component "experimental-istio-operator" {}`,
			expectedManifestName: testutil.ObjectMetadata{
				Version: "install.istio.io/v1alpha1", Kind: "IstioOperator", Name: "istiocontrolplane",
			},
			jsonPath: "{.spec.profile}",
			expected: "minimal",
		},
		{
			name: "demo profile",
			inputConfig: `component "experimental-istio-operator" {
				profile = "demo"
			}`,
			expectedManifestName: testutil.ObjectMetadata{
				Version: "install.istio.io/v1alpha1", Kind: "IstioOperator", Name: "istiocontrolplane",
			},
			jsonPath: "{.spec.profile}",
			expected: "demo",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			component := NewConfig()
			m := testutil.RenderManifests(t, component, Name, tc.inputConfig)
			gotConfig := testutil.ConfigFromMap(t, m, tc.expectedManifestName)

			testutil.MatchJSONPathStringValue(t, gotConfig, tc.jsonPath, tc.expected)
		})
	}
}

func TestVerifyServiceMonitor(t *testing.T) {
	inputConfig := `component "experimental-istio-operator" {
		enable_monitoring = true
	}`

	component := NewConfig()
	m := testutil.RenderManifests(t, component, Name, inputConfig)
	testutil.ConfigFromMap(t, m, testutil.ObjectMetadata{
		Version: "monitoring.coreos.com/v1", Kind: "ServiceMonitor", Name: "istio-operator",
	})
}

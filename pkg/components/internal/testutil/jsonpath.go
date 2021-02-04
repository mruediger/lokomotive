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

package testutil

import (
	"bufio"
	"io"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/util/jsonpath"
	yamlconv "sigs.k8s.io/yaml"
)

// unstructredObj accepts a Kubernetes manifest in YAML format and returns an object of type
// `unstructured.Unstructured`. This object has many methods that can be used by the consumer to
// extract metadata from the Kubernetes manifest.
func unstructredObj(t *testing.T, yamlObj string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}

	// Decode YAML into `unstructured.Unstructured`.
	dec := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if _, _, err := dec.Decode([]byte(yamlObj), nil, u); err != nil {
		t.Fatalf("Converting config to unstructured.Unstructured: %v", err)
	}

	return u
}

// removeYAMLComments converts YAML to JSON and back again, this removes the comments in the YAML
// and any extra whitespaces spaces.
func removeYAMLComments(t *testing.T, yamlObj []byte) []byte {
	jsonObj, err := yamlconv.YAMLToJSON(yamlObj)
	if err != nil {
		t.Fatalf("Converting YAML to JSON: %v", err)
	}

	yamlObj, err = yamlconv.JSONToYAML(jsonObj)
	if err != nil {
		t.Fatalf("Converting JSON to YAML: %v", err)
	}

	return yamlObj
}

// splitYAMLDocs converts a YAML string with multiple YAML docs separated by `---` into unique
// objects and returns those objects as a map.
func splitYAMLDocs(t *testing.T, yamlObj string) map[ObjectMetadata]string {
	ret := make(map[ObjectMetadata]string)

	reader := yaml.NewYAMLReader(bufio.NewReader(strings.NewReader(yamlObj)))

	for {
		yamlManifest, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("Error reading the YAML: %v", err)
		}

		yamlManifest = removeYAMLComments(t, yamlManifest)
		if string(yamlManifest) == "null\n" {
			continue
		}

		u := unstructredObj(t, string(yamlManifest))

		if u.GetAPIVersion() == "" || u.GetKind() == "" {
			t.Fatalf("invalid configuration no APIVersion or Kind: %s", string(yamlManifest))
		}

		obj := ObjectMetadata{
			Name:    u.GetName(),
			Kind:    u.GetKind(),
			Version: u.GetAPIVersion(),
		}

		ret[obj] = string(yamlManifest)
	}

	return ret
}

// valFromObject takes a JSON path as a string and an object of type `unstructured.Unstructured`.
// This function returns an object of type `reflect.Value` at that JSON path.
func valFromObject(t *testing.T, jp string, obj *unstructured.Unstructured) reflect.Value {
	jPath := jsonpath.New("parse")
	if err := jPath.Parse(jp); err != nil {
		t.Fatalf("Parsing JSONPath: %v", err)
	}

	v, err := jPath.FindResults(obj.Object)
	if err != nil {
		t.Fatalf("Finding results using JSONPath in the YAML file: %v", err)
	}

	if len(v) == 0 || len(v[0]) == 0 {
		t.Fatalf("No result found")
	}

	return v[0][0]
}

// jsonPathValue extracts an object at a JSON path from a YAML config, and returns an interface
// object.
func jsonPathValue(t *testing.T, yamlConfig string, jsonPath string) interface{} {
	u := unstructredObj(t, yamlConfig)
	got := valFromObject(t, jsonPath, u)

	switch got.Kind() { //nolint:exhaustive
	case reflect.Interface:
		// TODO: Add type switch here for concrete types.
		return got.Interface()
	default:
		t.Fatalf("Extracted object has an unknown type: %v", got.Kind())
	}

	return nil
}

// MatchJSONPathStringValue is a helper function for component unit tests. It compares the string at
// a JSON path in a YAML config to the expected string.
func MatchJSONPathStringValue(t *testing.T, yamlConfig string, jsonPath string, expected string) {
	obj := jsonPathValue(t, yamlConfig, jsonPath)

	got, ok := obj.(string)
	if !ok {
		t.Fatalf("Value is not string: %#v", obj)
	}

	if got != expected {
		t.Fatalf("Expected: %s, Got: %s", expected, got)
	}
}

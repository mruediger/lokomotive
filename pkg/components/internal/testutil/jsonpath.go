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
	"fmt"
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
func unstructredObj(yamlObj string) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}

	// Decode YAML into `unstructured.Unstructured`.
	dec := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if _, _, err := dec.Decode([]byte(yamlObj), nil, u); err != nil {
		return nil, fmt.Errorf("converting config to unstructured.Unstructured: %w", err)
	}

	return u, nil
}

// removeYAMLComments converts YAML to JSON and back again, this removes the comments in the YAML
// and any extra whitespaces spaces.
func removeYAMLComments(yamlObj []byte) ([]byte, error) {
	jsonObj, err := yamlconv.YAMLToJSON(yamlObj)
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	yamlObj, err = yamlconv.JSONToYAML(jsonObj)
	if err != nil {
		return nil, fmt.Errorf("converting JSON to YAML: %w", err)
	}

	return yamlObj, nil
}

// splitYAMLDocs converts a YAML string with multiple YAML docs separated by `---` into unique
// objects and returns those objects as a map.
func splitYAMLDocs(yamlObj string) (map[ObjectMetadata]string, error) {
	ret := make(map[ObjectMetadata]string)

	reader := yaml.NewYAMLReader(bufio.NewReader(strings.NewReader(yamlObj)))

	for {
		// Read the YAML document delimited by `---`.
		yamlManifest, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading the YAML: %w", err)
		}

		yamlManifest, err = removeYAMLComments(yamlManifest)
		if err != nil {
			return nil, fmt.Errorf("removing YAML comments: %w", err)
		}

		// Check if the YAML is empty.
		if string(yamlManifest) == "null\n" {
			continue
		}

		u, err := unstructredObj(string(yamlManifest))
		if err != nil {
			return nil, fmt.Errorf("YAML to unstructured object: %w", err)
		}

		if u.GetAPIVersion() == "" || u.GetKind() == "" {
			return nil, fmt.Errorf("invalid configuration no APIVersion or Kind: %s", string(yamlManifest))
		}

		obj := ObjectMetadata{
			Name:    u.GetName(),
			Kind:    u.GetKind(),
			Version: u.GetAPIVersion(),
		}

		ret[obj] = string(yamlManifest)
	}

	return ret, nil
}

// valFromObject takes a JSON path as a string and an object of type `unstructured.Unstructured`.
// This function returns an object of type `reflect.Value` at that JSON path.
func valFromObject(jp string, obj *unstructured.Unstructured) (reflect.Value, error) {
	jPath := jsonpath.New("parse")
	if err := jPath.Parse(jp); err != nil {
		return reflect.Value{}, fmt.Errorf("parsing JSONPath: %w", err)
	}

	v, err := jPath.FindResults(obj.Object)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("finding results using JSONPath in the YAML file: %w", err)
	}

	if len(v) == 0 || len(v[0]) == 0 {
		return reflect.Value{}, fmt.Errorf("no result found")
	}

	return v[0][0], nil
}

// jsonPathValue extracts an object at a JSON path from a YAML config, and returns an interface
// object.
func jsonPathValue(yamlConfig string, jsonPath string) (interface{}, error) {
	u, err := unstructredObj(yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("YAML to unstructured object: %w", err)
	}

	got, err := valFromObject(jsonPath, u)
	if err != nil {
		return nil, fmt.Errorf("JSON path value in YAML: %w", err)
	}

	switch got.Kind() { //nolint:exhaustive
	case reflect.Interface:
		// TODO: Add type switch here for concrete types.
		return got.Interface(), nil
	default:
		return nil, fmt.Errorf("extracted object has an unknown type: %v", got.Kind())
	}
}

// MatchJSONPathStringValue is a helper function for component unit tests. It compares the string at
// a JSON path in a YAML config to the expected string.
func MatchJSONPathStringValue(t *testing.T, yamlConfig string, jsonPath string, expected string) {
	obj, err := jsonPathValue(yamlConfig, jsonPath)
	if err != nil {
		t.Fatalf("Extracting JSON path value: %v", err)
	}

	got, ok := obj.(string)
	if !ok {
		t.Fatalf("Value is not string: %#v", obj)
	}

	if got != expected {
		t.Fatalf("Expected: %s, Got: %s", expected, got)
	}
}

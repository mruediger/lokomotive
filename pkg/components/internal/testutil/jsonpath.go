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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

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
	u, err := util.YAMLToUnstructured([]byte(yamlConfig))
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

// JSONPathExists checks if the given YAML config has an object at the given JSON path, also provide
// what error to expect.
func JSONPathExists(t *testing.T, yamlConfig string, jsonPath string, errExp string) {
	_, err := jsonPathValue(yamlConfig, jsonPath)
	if err != nil && errExp == "" { //nolint:gocritic
		t.Fatalf("Error not expected and failed with: %v", err)
	} else if err != nil && !strings.Contains(err.Error(), errExp) {
		t.Fatalf("Extracting JSON path value, expected error: %v to contain: %q", err, errExp)
	} else if err == nil && errExp != "" {
		t.Fatalf("Expected error %q but got none", errExp)
	}
}

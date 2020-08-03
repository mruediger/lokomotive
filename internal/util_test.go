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
package internal_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/internal"
)

const namespace = "test"

func TestApppendNamespaceLabelSuccess(t *testing.T) {
	m := map[string]string{
		"key": "value",
	}

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

func TestApppendNamespaceLabelDontAllowOverride(t *testing.T) {
	value := "old_value"
	m := map[string]string{
		internal.NamespaceLabelKey: value,
	}

	final := internal.AppendNamespaceLabel(namespace, m)

	if final[internal.NamespaceLabelKey] == namespace {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, value)
	}
}

func TestApppendNamespaceLabelNilMap(t *testing.T) {
	var m map[string]string

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

func TestApppendNamespaceLabelEmptyMap(t *testing.T) {
	m := map[string]string{}

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

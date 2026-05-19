package chart_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestManagerRBACMatchesControllerGen(t *testing.T) {
	repoRoot := repoRoot(t)
	rendered := run(t, repoRoot,
		"helm", "template", "template-k8s", "charts/template-k8s",
		"--namespace", "template-k8s-system",
		"--show-only", "templates/rbac-manager.yaml",
	)

	chartRole := findObject(t, rendered, "ClusterRole", "template-k8s-manager-role")
	generatedRoleDir := filepath.Join(t.TempDir(), "rbac")
	run(t, repoRoot,
		"controller-gen", "rbac:roleName=manager-role", "paths=./...",
		"output:rbac:dir="+generatedRoleDir,
	)
	generatedRole := readObject(t, filepath.Join(generatedRoleDir, "role.yaml"))

	if got, want := canonicalRules(t, chartRole), canonicalRules(t, generatedRole); got != want {
		t.Fatalf("chart manager RBAC drifted from controller-gen output\nchart: %s\ncontroller-gen: %s", got, want)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root")
		}
		dir = parent
	}
}

func run(t *testing.T, dir string, name string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
	return out
}

func readObject(t *testing.T, path string) *unstructured.Unstructured {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	objects := decodeObjects(t, data)
	if len(objects) != 1 {
		t.Fatalf("expected one object in %s, got %d", path, len(objects))
	}
	return objects[0]
}

func findObject(t *testing.T, data []byte, kind string, name string) *unstructured.Unstructured {
	t.Helper()

	for _, obj := range decodeObjects(t, data) {
		if obj.GetKind() == kind && obj.GetName() == name {
			return obj
		}
	}
	t.Fatalf("could not find %s/%s", kind, name)
	return nil
}

func decodeObjects(t *testing.T, data []byte) []*unstructured.Unstructured {
	t.Helper()

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	var objects []*unstructured.Unstructured
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(obj.Object) == 0 {
			continue
		}
		objects = append(objects, obj)
	}
	return objects
}

func canonicalRules(t *testing.T, obj *unstructured.Unstructured) string {
	t.Helper()

	rawRules, ok, err := unstructured.NestedSlice(obj.Object, "rules")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("%s/%s has no rules", obj.GetKind(), obj.GetName())
	}

	rules := make([]string, 0, len(rawRules))
	for _, rawRule := range rawRules {
		rule, ok := rawRule.(map[string]any)
		if !ok {
			t.Fatalf("unexpected RBAC rule type %T", rawRule)
		}
		normalized := map[string][]string{}
		for _, key := range []string{"apiGroups", "resources", "verbs", "nonResourceURLs"} {
			if values, ok := rule[key]; ok {
				normalized[key] = sortedStrings(t, values)
			}
		}
		data, err := json.Marshal(normalized)
		if err != nil {
			t.Fatal(err)
		}
		rules = append(rules, string(data))
	}
	sort.Strings(rules)

	data, err := json.Marshal(rules)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func sortedStrings(t *testing.T, values any) []string {
	t.Helper()

	items, ok := values.([]any)
	if !ok {
		t.Fatalf("unexpected RBAC field type %T", values)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("unexpected RBAC string value type %T", item)
		}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

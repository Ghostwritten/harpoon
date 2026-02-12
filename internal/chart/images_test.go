package chart

import (
	"testing"
)

func TestExtractImagesFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantNum int
		wantHas []string
	}{
		{
			name: "single doc with containers",
			yaml: `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: nginx
    image: nginx:1.21-alpine
  - name: sidecar
    image: quay.io/foo/bar:v1
`,
			wantNum: 2,
			wantHas: []string{"nginx:1.21-alpine", "quay.io/foo/bar:v1"},
		},
		{
			name: "initContainers and containers",
			yaml: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      initContainers:
      - name: init
        image: busybox:1.36
      containers:
      - name: app
        image: myreg.io/proj/app:latest
`,
			wantNum: 2,
			wantHas: []string{"busybox:1.36", "myreg.io/proj/app:latest"},
		},
		{
			name: "multi-document stream",
			yaml: `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg
---
apiVersion: v1
kind: Pod
spec:
  containers:
  - image: nginx:latest
    name: nginx
---
apiVersion: v1
kind: Pod
spec:
  containers:
  - image: nginx:latest
    name: nginx2
`,
			wantNum: 1,
			wantHas: []string{"nginx:latest"},
		},
		{
			name: "nested image keys",
			yaml: `
spec:
  image: registry.example.com/operator:2.0
  config:
    image: busybox:1.36
`,
			wantNum: 2,
			wantHas: []string{"registry.example.com/operator:2.0", "busybox:1.36"},
		},
		{
			name: "empty and no image",
			yaml: `
apiVersion: v1
kind: Namespace
metadata:
  name: foo
`,
			wantNum: 0,
			wantHas: nil,
		},
		{
			name: "image embedded in args (e.g. prometheus-operator config-reloader)",
			yaml: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: operator
        args:
        - --prometheus-config-reloader=quay.io/prometheus-operator/prometheus-config-reloader:v0.88.0
`,
			wantNum: 1,
			wantHas: []string{"quay.io/prometheus-operator/prometheus-config-reloader:v0.88.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractImagesFromYAML([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("ExtractImagesFromYAML: %v", err)
			}
			if len(got) != tt.wantNum {
				t.Errorf("got %d images, want %d: %v", len(got), tt.wantNum, got)
			}
			gotSet := make(map[string]struct{})
			for _, s := range got {
				gotSet[s] = struct{}{}
			}
			for _, w := range tt.wantHas {
				if _, ok := gotSet[w]; !ok {
					t.Errorf("missing wanted image %q, got: %v", w, got)
				}
			}
		})
	}
}

func TestExtractImagesFromYAML_InvalidYAML(t *testing.T) {
	_, err := ExtractImagesFromYAML([]byte("invalid: yaml: ["))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestExtractImagesFromYAML_RejectURLsAndMetrics(t *testing.T) {
	// URLs and Prometheus metric names must not appear as images
	yaml := `
spec:
  runbook_url: https://runbooks.prometheus-operator.dev/runbooks/alertmanager/alertmanagerclusterdown
  expr: instance:node_cpu:ratio
  targets:
  - http://localhost:3000/api/reload
containers:
- image: quay.io/prometheus/prometheus:v3.9.1
  args:
  - --config.reload.url=http://localhost:9090/-/reload
`
	got, err := ExtractImagesFromYAML([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "quay.io/prometheus/prometheus:v3.9.1" {
		t.Errorf("expected only quay.io/prometheus/prometheus:v3.9.1, got %v", got)
	}
}

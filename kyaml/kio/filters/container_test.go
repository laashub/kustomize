// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestContainerFilter_Filter(t *testing.T) {
	var tests = []struct {
		name              string
		input             []string
		expectedOutput    []string
		expectedError     string
		expectedResults   string
		noMakeResultsFile bool
		instance          ContainerFilter
	}{
		{
			name: "add_path_annotation",
			instance: ContainerFilter{args: []string{
				"echo", `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
`,
			},
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
		},

		{
			name: "write_results",
			instance: ContainerFilter{args: []string{
				"echo", `
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:  
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
			},
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
		},

		{
			name:          "write_results_non_0_exit",
			expectedError: "exit status 1",
			instance: ContainerFilter{args: []string{"sh", "-c",
				`echo '
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:  
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
' && cat not-real-dir
`,
			},
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
		},

		{
			name:              "write_results_non_0_exit_missing_file",
			expectedError:     "open /not/real/file: no such file or directory",
			noMakeResultsFile: true,
			instance: ContainerFilter{args: []string{"sh", "-c",
				`echo '
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
results:
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:  
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
' && cat not-real-dir
`,
			},
			},
			expectedOutput: []string{
				`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
`,
				`
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'service_service-foo.yaml'
`,
			},
			expectedResults: `
- apiVersion: config.k8s.io/v1alpha1
  kind: ObjectError
  name: "some-validator"
  items:
  - type: error
    message: "some message"
    resourceRef:
      apiVersion: apps/v1
      kind: Deployment
      name: foo
      namespace: bar
    file:
      path: deploy.yaml
      index: 0
    field:
      path: "spec.template.spec.containers[3].resources.limits.cpu"
      currentValue: "200"
      suggestedValue: "2"
`,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expectedResults) > 0 && !tt.noMakeResultsFile {
				f, err := ioutil.TempFile("", "test-kyaml-*.yaml")
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				defer os.RemoveAll(f.Name())
				tt.instance.ResultsFile = f.Name()
			} else if len(tt.expectedResults) > 0 {
				tt.instance.ResultsFile = "/not/real/file"
			}

			var inputs []*yaml.RNode
			for i := range tt.input {
				node, err := yaml.Parse(tt.input[i])
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inputs = append(inputs, node)
			}

			output, err := tt.instance.Filter(inputs)
			if tt.expectedError != "" {
				if !assert.EqualError(t, err, tt.expectedError) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			var actual []string
			for i := range output {
				s, err := output[i].String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				actual = append(actual, strings.TrimSpace(s))
			}
			var expected []string
			for i := range tt.expectedOutput {
				expected = append(expected, strings.TrimSpace(tt.expectedOutput[i]))
			}

			if !assert.Equal(t, expected, actual) {
				t.FailNow()
			}

			if len(tt.instance.ResultsFile) > 0 {
				tt.expectedResults = strings.TrimSpace(tt.expectedResults)

				results, err := tt.instance.Results.String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.expectedResults, strings.TrimSpace(results)) {
					t.FailNow()
				}

				b, err := ioutil.ReadFile(tt.instance.ResultsFile)
				writtenResults := strings.TrimSpace(string(b))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.expectedResults, writtenResults) {
					t.FailNow()
				}
			}
		})
	}
}

func TestFilter_command(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	instance := &ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
	}
	os.Setenv("KYAML_TEST", "FOO")
	cmd := instance.getCommand()

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "none",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
	}
	for _, e := range os.Environ() {
		// the process env
		expected = append(expected, "-e", strings.Split(e, "=")[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)

	foundKyaml := false
	for _, e := range cmd.Env {
		// verify the command has the right environment variables to pass to the container
		split := strings.Split(e, "=")
		if split[0] == "KYAML_TEST" {
			assert.Equal(t, "FOO", split[1])
			foundKyaml = true
		}
	}
	assert.True(t, foundKyaml)
}

func TestFilter_command_StorageMount(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	bindMount := StorageMount{"bind", "/mount/path", "/local/"}
	localVol := StorageMount{"volume", "myvol", "/local/"}
	tmpfs := StorageMount{"tmpfs", "", "/local/"}
	instance := &ContainerFilter{
		Image:         "example.com:version",
		Config:        cfg,
		StorageMounts: []StorageMount{bindMount, localVol, tmpfs},
	}
	cmd := instance.getCommand()

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "none",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "bind", "/mount/path", "/local/"),
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "volume", "myvol", "/local/"),
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "tmpfs", "", "/local/"),
	}
	for _, e := range os.Environ() {
		// the process env
		expected = append(expected, "-e", strings.Split(e, "=")[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)
}

func TestFilter_command_network(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	instance := &ContainerFilter{
		Image:   "example.com:version",
		Network: "test-net",
		Config:  cfg,
	}
	cmd := instance.getCommand()

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "test-net",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
	}
	for _, e := range os.Environ() {
		// the process env
		tokens := strings.Split(e, "=")
		if tokens[0] == "" {
			continue
		}
		expected = append(expected, "-e", tokens[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)
}

func TestFilter_Filter(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'statefulset_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_noChange(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sh", "-c", "cat <&0"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'service_service-foo.yaml'
`, b.String())
}

func Test_GetFunction(t *testing.T) {
	var tests = []struct {
		name       string
		resource   string
		expectedFn string
		missingFn  bool
	}{

		// fn annotation
		{
			name: "fn annotation",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
`,
			expectedFn: `
container:
    image: foo:v1.0.0`,
		},

		{
			name: "storage mounts json style",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        mounts: [ {type: bind, src: /mount/path, dst: /local/}, {src: myvol, dst: /local/, type: volume}, {dst: /local/, type: tmpfs} ]
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    mounts:
      - type: bind
        src: /mount/path
        dst: /local/
      - type: volume
        src: myvol
        dst: /local/
      - type: tmpfs
        dst: /local/
`,
		},

		{
			name: "storage mounts yaml style",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        mounts:
        - src: /mount/path
          type: bind
          dst: /local/
        - dst: /local/
          src: myvol
          type: volume
        - type: tmpfs
          dst: /local/
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    mounts:
      - type: bind
        src: /mount/path
        dst: /local/
      - type: volume
        src: myvol
        dst: /local/
      - type: tmpfs
        dst: /local/
`,
		},

		{
			name: "network",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
        network:
          required: true
`,
			expectedFn: `
container:
    image: foo:v1.0.0
    network:
        required: true
`,
		},

		{
			name: "path",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      path: foo
      container:
        image: foo:v1.0.0
`,
			// path should be erased
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		{
			name: "network",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations:
    config.kubernetes.io/function: |-
      network: foo
      container:
        image: foo:v1.0.0
`,
			// network should be erased
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		// legacy fn style
		{name: "legacy fn meta",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  configFn:
      container:
        image: foo:v1.0.0
`,
			expectedFn: `
container:
    image: foo:v1.0.0
`,
		},

		// no fn
		{name: "no fn",
			resource: `
apiVersion: v1beta1
kind: Example
metadata:
  annotations: {}
`,
			missingFn: true,
		},

		// test network, etc...
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			resource := yaml.MustParse(tt.resource)
			fn := GetFunctionSpec(resource)
			if tt.missingFn {
				if !assert.Nil(t, fn) {
					t.FailNow()
				}
			} else {
				b, err := yaml.Marshal(fn)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tt.expectedFn),
					strings.TrimSpace(string(b))) {
					t.FailNow()
				}
			}
		})
	}
}

func Test_GetContainerNetworkRequired(t *testing.T) {
	tests := []struct {
		input    string
		required bool
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      network:
        required: true
`,
			required: true,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
      network:
        required: false
`,
			required: false,
		},
		{

			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  configFn:
    container:
      image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
`,
			required: false,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kustomize-functions/example-tshirt:v0.1.0
        network:
          required: true
`,
			required: true,
		},
	}

	for _, tc := range tests {
		cfg, err := yaml.Parse(tc.input)
		if !assert.NoError(t, err) {
			return
		}
		fn := GetFunctionSpec(cfg)
		assert.Equal(t, tc.required, fn.Container.Network.Required)
	}
}

func TestFilter_Filter_defaultNaming(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(``)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args: []string{"echo", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_defaultNamingFunctions(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(``)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args: []string{"echo", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/functions/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_scopeMissingFromResource(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved -- paths shouldn't be set by container
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_globalScope(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		GlobalScope:           true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved -- paths shouldn't be set by container
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/statefulset_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_scopeFunctionsDir(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/functions/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be modified
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_scope_nested_resource(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'baz.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'baz.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be modified
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_scopeDir(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		SetFlowStyleForConfig: true,
		Image:                 "example.com:version",
		Config:                cfg,
		args:                  []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestContainerFilter_scope(t *testing.T) {
	cf := &ContainerFilter{}

	fnR, err := yaml.Parse(`apiVersion: config.kubernetes.io/v1beta1
kind: ConfigFunction
metadata:
  name: config-function
  annotations:
    config.kubernetes.io/path: 'functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	inRs := []*yaml.RNode{fnR}
	inScopeRs, notInScopeRs, err := cf.scope(".", inRs)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, inScopeRs, 1, "Number of in-scope Resources")
	assert.Len(t, notInScopeRs, 0, "Number of out-of-scope Resources")
}

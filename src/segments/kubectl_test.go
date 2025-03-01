package segments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

const (
	testKubectlAllInfoTemplate = "{{.Context}} :: {{.Namespace}} :: {{.User}} :: {{.Cluster}}"
	contextMarker              = "currentcontextmarker"
)

func TestKubectlSegment(t *testing.T) {
	standardTemplate := "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}"
	lsep := string(filepath.ListSeparator)

	cases := []struct {
		Files           map[string]string
		ContextAliases  map[string]string
		Cluster         string
		Kubeconfig      string
		Context         string
		Namespace       string
		UserName        string
		Case            string
		ExpectedString  string
		Template        string
		KubectlExists   bool
		ParseKubeConfig bool
		KubectlErr      bool
		ExpectedEnabled bool
		DisplayError    bool
	}{
		{
			Case:            "kubeconfig incomplete",
			Template:        testKubectlAllInfoTemplate,
			ParseKubeConfig: true,
			Kubeconfig:      contextMarker + lsep + "contextdefinitionincomplete",
			Files:           testKubeConfigFiles,
			ExpectedString:  "ctx ::  ::  ::",
			ExpectedEnabled: true,
		},
		{Case: "disabled", Template: standardTemplate, KubectlExists: false, Context: "aaa", Namespace: "bbb", ExpectedEnabled: false},
		{
			Case:            "all information",
			Template:        testKubectlAllInfoTemplate,
			KubectlExists:   true,
			Context:         "aaa",
			Namespace:       "bbb",
			UserName:        "ccc",
			Cluster:         "ddd",
			ExpectedString:  "aaa :: bbb :: ccc :: ddd",
			ExpectedEnabled: true,
		},
		{Case: "no namespace", Template: standardTemplate, KubectlExists: true, Context: "aaa", ExpectedString: "aaa", ExpectedEnabled: true},
		{
			Case:            "kubectl context alias",
			Template:        standardTemplate,
			KubectlExists:   true,
			Context:         "aaa",
			Namespace:       "bbb",
			ContextAliases:  map[string]string{"aaa": "ccc"},
			ExpectedString:  "ccc :: bbb",
			ExpectedEnabled: true,
		},
		{
			Case:            "kubectl error",
			Template:        standardTemplate,
			DisplayError:    true,
			KubectlExists:   true,
			Context:         "aaa",
			Namespace:       "bbb",
			KubectlErr:      true,
			ExpectedString:  "KUBECTL ERR :: KUBECTL ERR",
			ExpectedEnabled: true,
		},
		{Case: "kubectl error hidden", Template: standardTemplate, DisplayError: false, KubectlExists: true, Context: "aaa", Namespace: "bbb", KubectlErr: true, ExpectedEnabled: false},
		{
			Case:            "kubeconfig home",
			Template:        testKubectlAllInfoTemplate,
			ParseKubeConfig: true,
			Files:           testKubeConfigFiles,
			ExpectedString:  "aaa :: bbb :: ccc :: ddd",
			ExpectedEnabled: true,
		},
		{
			Case:            "kubeconfig context alias",
			Template:        standardTemplate,
			ParseKubeConfig: true,
			Files:           testKubeConfigFiles,
			ContextAliases:  map[string]string{"aaa": "ccc"},
			ExpectedString:  "ccc :: bbb",
			ExpectedEnabled: true,
		},
		{
			Case:            "kubeconfig multiple current marker first",
			Template:        testKubectlAllInfoTemplate,
			ParseKubeConfig: true,
			Kubeconfig:      "" + lsep + contextMarker + lsep + "contextdefinition" + lsep + "contextredefinition",
			Files:           testKubeConfigFiles,
			ExpectedString:  "ctx :: ns :: usr :: cl",
			ExpectedEnabled: true,
		},
		{
			Case:     "kubeconfig multiple context first",
			Template: testKubectlAllInfoTemplate, ParseKubeConfig: true,
			Kubeconfig:      "contextdefinition" + lsep + "contextredefinition" + lsep + contextMarker + lsep,
			Files:           testKubeConfigFiles,
			ExpectedString:  "ctx :: ns :: usr :: cl",
			ExpectedEnabled: true,
		},
		{
			Case: "kubeconfig error hidden", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true, Kubeconfig: "invalid", Files: testKubeConfigFiles, ExpectedEnabled: false},
		{
			Case:            "kubeconfig error",
			Template:        testKubectlAllInfoTemplate,
			ParseKubeConfig: true,
			Kubeconfig:      "invalid",
			Files:           testKubeConfigFiles,
			DisplayError:    true,
			ExpectedString:  "KUBECONFIG ERR :: KUBECONFIG ERR :: KUBECONFIG ERR :: KUBECONFIG ERR",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "kubectl").Return(tc.KubectlExists)

		var kubeconfig string
		content, err := os.ReadFile("../test/kubectl.yml")
		if err == nil {
			kubeconfig = fmt.Sprintf(string(content), tc.Cluster, tc.UserName, tc.Namespace, tc.Context)
		}

		var kubectlErr error
		if tc.KubectlErr {
			kubectlErr = &runtime.CommandError{
				Err:      "oops",
				ExitCode: 1,
			}
		}

		env.On("RunCommand", "kubectl", []string{"config", "view", "--output", "yaml", "--minify"}).Return(kubeconfig, kubectlErr)
		env.On("Getenv", "KUBECONFIG").Return(tc.Kubeconfig)

		for path, content := range tc.Files {
			env.On("FileContent", path).Return(content)
		}

		env.On("Home").Return("testhome")

		props := properties.Map{
			properties.DisplayError: tc.DisplayError,
			ParseKubeConfig:         tc.ParseKubeConfig,
			ContextAliases:          tc.ContextAliases,
		}

		k := &Kubectl{}
		k.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, k.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, k), tc.Case)
		}
	}
}

var testKubeConfigFiles = map[string]string{
	filepath.Join("testhome", ".kube/config"): `
apiVersion: v1
contexts:
  - context:
      cluster: ddd
      user: ccc
      namespace: bbb
    name: aaa
current-context: aaa
`,
	"contextdefinition": `
apiVersion: v1
contexts:
  - context:
      cluster: cl
      user: usr
      namespace: ns
    name: ctx
`,
	contextMarker: `
apiVersion: v1
current-context: ctx
`,
	"invalid": "this is not yaml",
	"contextdefinitionincomplete": `
apiVersion: v1
contexts:
  - name: ctx
`,
	"contextredefinition": `
apiVersion: v1
contexts:
  - context:
      cluster: wrongcl
      user: wrongu
      namespace: wrongns
    name: ctx
`,
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
)

var (
	kubectl = flag.String("kubectl", "kubectl", "kubectl command")
)

func main() {
	flag.Parse()
	configs := listConfigs()
	displayAndChooseConfig(configs)
}

type kubernetesConfig struct {
	Contexts       []kubernetesContextEntry `json:"contexts"`
	CurrentContext string            `json:"current-context"`
}

type kubernetesContextEntry struct {
	Name    string     `json:"name"`
	Context kubernetesContext `json:"context"`
}

type kubernetesContext struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

func listConfigs() kubernetesConfig {
	b, err := exec.Command(*kubectl, "config", "view", "-ojson").CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("listing configurations: %w", err))
	}
	var config kubernetesConfig
	if err := json.Unmarshal(b, &config); err != nil {
		panic(fmt.Errorf("json unmarshalling bytes: %q, %w", string(b), err))
	}
	return config
}

func useConfig(c kubernetesContextEntry) []byte {
	b, err := exec.Command(*kubectl, "config", "use-context", c.Name).CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("activating configuration: %q, %w", string(b), err))
	}
	return b
}

func displayAndChooseConfig(config kubernetesConfig) {
	var activeIndex int
	for i, c := range config.Contexts {
		if c.Name == config.CurrentContext {
			activeIndex = i
			break
		}
	}

	f := template.FuncMap{}
	for k, v := range promptui.FuncMap {
		f[k] = v
	}
	f["IsCurrent"] = func(c kubernetesContextEntry) bool {
		return c.Name == config.CurrentContext
	}
	isGke := func(c kubernetesContextEntry) bool {
		return strings.HasPrefix(c.Name, "gke_") && len(strings.Split(c.Name, "_")) == 4
	}
	f["IsGke"] = isGke
	f["GkeProject"] = func(c kubernetesContextEntry) string {
		if !isGke(c) {
			return "ERROR"
		}
		return strings.Split(c.Name, "_")[1]
	}
	f["GkeLocation"] = func(c kubernetesContextEntry) string {
		if !isGke(c) {
			return "ERROR"
		}
		return strings.Split(c.Name, "_")[2]
	}
	f["GkeCluster"] = func(c kubernetesContextEntry) string {
		if !isGke(c) {
			return "ERROR"
		}
		return strings.Split(c.Name, "_")[3]
	}

	prompt := promptui.Select{
		Label: "kubectl config get-contexts",
		Items: config.Contexts,
		Templates: &promptui.SelectTemplates{
			Active: "{{ .Name | cyan | underline }}" +
				"{{ if IsCurrent . }} {{- \" (Active)\" | cyan | underline }} {{ end }}",
			Inactive: "{{ .Name }}{{ if IsCurrent . }} {{- \" (Active)\" }} {{ end }}",
			Details: "{{ if IsGke . }}" +
				"Project: {{ GkeProject . }}" +
				"\tLocation: {{ GkeLocation . }}" +
				"\tCluster: {{ GkeCluster . }}" +
				"{{ end }}",
			FuncMap: f,
		},
		Size: len(config.Contexts),
		HideSelected: true,
	}
	i, _, err := prompt.RunCursorAt(activeIndex, 0)
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}
	b := useConfig(config.Contexts[i])
	fmt.Print(string(b))
}

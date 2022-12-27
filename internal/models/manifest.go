package models

type Manifest struct {
	ApiVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
	Data       map[string]interface{} `yaml:"data"`
	StringData map[string]interface{} `yaml:"stringData"`
}

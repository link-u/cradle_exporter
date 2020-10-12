package cradle

import (
	"io/ioutil"

	"github.com/immortal/immortal"
	"gopkg.in/yaml.v2"
)

type ExporterConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

type ScriptConfig struct {
	Path string `yaml:"path,omitempty"`
}

type CronConfig struct {
	Path string `yaml:"path,omitempty"`
}

type StaticConfig struct {
	Path string `yaml:"path,omitempty"`
}

type TargetConfig struct {
	ExporterConfig   *ExporterConfig  `yaml:"exporter,omitempty"`
	SupervisorConfig *immortal.Config `yaml:"immortal,omitempty"`
	ScriptConfig     *ScriptConfig    `yaml:"script,omitempty"`
	CronConfig       *CronConfig      `yaml:"cron,omitempty"`
	StaticConfig     *StaticConfig    `yaml:"static,omitempty"`
}

type Config struct {
	IncludeDirs []string `yaml:"include_dirs,omitempty"`
}

func ReadTargetConfigFromFile(path string) (*TargetConfig, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadTargetConfig(bytes)
}

func ReadTargetConfig(bytes []byte) (*TargetConfig, error) {
	var err error
	var config TargetConfig
	err = yaml.UnmarshalStrict(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func ReadConfigFromFile(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadConfig(bytes)
}

func ReadConfig(bytes []byte) (*Config, error) {
	var err error
	var config Config
	err = yaml.UnmarshalStrict(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

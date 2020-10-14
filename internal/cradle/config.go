package cradle

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ExporterConfig struct {
	Endpoints []string `yaml:"endpoints,omitempty"`
}

type ScriptConfig struct {
	Path string   `yaml:"path,omitempty"`
	Args []string `yaml:"args,omitempty"`
}

type ServiceConfig struct {
	Path      string   `yaml:"path,omitempty"`
	Args      []string `yaml:"args,omitempty"`
	Endpoints []string `yaml:"endpoints,omitempty"`
}

type CronJobConfig struct {
	Path  string   `yaml:"path,omitempty"`
	Args  []string `yaml:"args,omitempty"`
	Every string   `yaml:"every,omitempty"`
}

type StaticFileConfig struct {
	Paths []string `yaml:"paths,omitempty"`
}

type TargetConfig struct {
	ConfigFilePath string            `yaml:",omitempty"`
	ExporterConfig *ExporterConfig   `yaml:"exporter,omitempty"`
	ServiceConfig  *ServiceConfig    `yaml:"service,omitempty"`
	ScriptConfig   *ScriptConfig     `yaml:"script,omitempty"`
	CronJobConfig  *CronJobConfig    `yaml:"cron,omitempty"`
	StaticConfig   *StaticFileConfig `yaml:"static,omitempty"`
}

type Config struct {
	IncludeDirs []string `yaml:"include_dirs,omitempty"`
}

func ReadTargetConfigFromFile(path string) (*TargetConfig, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, err := ReadTargetConfig(bytes)
	if config != nil {
		config.ConfigFilePath = path
	}
	return config, err
}

func ReadTargetConfig(bytes []byte) (*TargetConfig, error) {
	var err error
	var config TargetConfig
	config.ConfigFilePath = "<mem>"
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

func collectTargetConfigsFromDir(dpath string, dst map[string]*TargetConfig) error {
	dpath = filepath.Clean(dpath)
	info, err := os.Lstat(dpath)
	if err != nil {
		return err
	}
	if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
		dpath, err = os.Readlink(dpath)
		if err != nil {
			return err
		}
		dpath = filepath.Clean(dpath)
		info, err = os.Lstat(dpath)
		if err != nil {
			return err
		}
	}
	if !info.Mode().IsDir() {
		return fmt.Errorf("config dir is not dir: %o", info.Mode())
	}
	return filepath.Walk(dpath, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
			fpath, err = os.Readlink(fpath)
			if err != nil {
				return err
			}
			fpath = filepath.Clean(fpath)
			info, err = os.Lstat(fpath)
			if err != nil {
				return err
			}
		}
		config, err := ReadTargetConfigFromFile(fpath)
		if err != nil {
			return err
		}
		dst[fpath] = config
		return nil
	})
}

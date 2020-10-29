package cradle

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func newTargets(config *Config) (map[string]Target, error) {
	configs := make(map[string]*TargetConfig)
	for _, dir := range config.IncludeDirs {
		if err := collectTargetConfigsFromDir(dir, configs); err != nil {
			return nil, err
		}
	}
	targets := make(map[string]Target)
	for fpath, cfg := range configs {
		target := newTarget(cfg)
		if target == nil {
			yamlBytes, err := yaml.Marshal(cfg)
			if err != nil {
				return nil, fmt.Errorf("invalid config(unknown target type): \n%v", err)
			}
			return nil, fmt.Errorf("invalid config(unknown target type): \n%s", string(yamlBytes))
		}
		targets[fpath] = target
	}
	return targets, nil
}

//---

func newTarget(cfg *TargetConfig) Target {
	switch {
	case cfg.StaticConfig != nil:
		return &StaticFileTarget{
			Config: cfg,
		}
	case cfg.CronJobConfig != nil:
		return &CronJobTarget{
			Config: cfg,
		}
	case cfg.ScriptConfig != nil:
		return &ScriptTarget{
			Config: cfg,
		}
	case cfg.ServiceConfig != nil:
		return &ServiceTarget{
			Config: cfg,
		}
	case cfg.ExporterConfig != nil:
		return &ExporterTarget{
			Config: cfg,
		}
	default:
		return nil
	}
}

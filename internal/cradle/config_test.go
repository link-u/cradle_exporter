package cradle

import (
	"reflect"
	"testing" // テストで使える関数・構造体が用意されているパッケージをimport
)

func TestReadExporterConfig(t *testing.T) {
	const kConfigString = `
---
exporter:
  endpoints:
    - "https://example.com/"
`
	conf, err := ReadTargetConfig([]byte(kConfigString))
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	if conf.CronJobConfig != nil {
		t.Error("Config does not include cron config")
	}
	if conf.ScriptConfig != nil {
		t.Error("Config does not include script config")
	}
	if conf.StaticConfig != nil {
		t.Error("Config does not include static config")
	}
	if conf.ServiceConfig != nil {
		t.Error("Config does not include supervisor config")
	}
	expectedEndpoints := []string{"https://example.com/"}
	if !reflect.DeepEqual(conf.ExporterConfig.Endpoints, expectedEndpoints) {
		t.Errorf("Endpoint does not match: %v != %v", conf.ExporterConfig.Endpoints, expectedEndpoints)
	}
}

func TestReadScriptConfig(t *testing.T) {
	const kConfigString = `
---
script:
  path: /usr/bin/script
  args:
   - 'arg1'
   - 'arg2'
   - 'arg3'
`
	conf, err := ReadTargetConfig([]byte(kConfigString))
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	if conf.CronJobConfig != nil {
		t.Error("Config should not include cron config")
	}
	expectedPath := "/usr/bin/script"
	if conf.ScriptConfig.Path != expectedPath {
		t.Errorf("Script path does not match: %v != %v", conf.ScriptConfig.Path, expectedPath)
	}
	expectedArgs := []string{"arg1", "arg2", "arg3"}
	if !reflect.DeepEqual(conf.ScriptConfig.Args, expectedArgs) {
		t.Errorf("Script arg does not match: %v != %v", conf.ScriptConfig.Args, expectedArgs)
	}
	if conf.StaticConfig != nil {
		t.Error("Config should not include static config")
	}
	if conf.ServiceConfig != nil {
		t.Error("Config should not include supervisor config")
	}
	if conf.ExporterConfig != nil {
		t.Error("Config should not include exporter config")
	}
}

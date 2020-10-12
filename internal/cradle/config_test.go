package cradle

import (
	"testing" // テストで使える関数・構造体が用意されているパッケージをimport
)

func TestReadTargetConfig(t *testing.T) {
	const kConfigString = `
---
exporter:
  endpoint: "https://example.com/"
`
	conf, err := ReadTargetConfig([]byte(kConfigString))
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	if conf.CronConfig != nil {
		t.Error("Config does not include cron config")
	}
	if conf.ScriptConfig != nil {
		t.Error("Config does not include script config")
	}
	if conf.StaticConfig != nil {
		t.Error("Config does not include static config")
	}
	if conf.SupervisorConfig != nil {
		t.Error("Config does not include supervisor config")
	}
	if conf.ExporterConfig.Endpoint != "https://example.com/" {
		t.Errorf("Endpoint does not match: %v != %v", conf.ExporterConfig.Endpoint, "https://example.com/")
	}
}

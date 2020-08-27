package log

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger("test1")
	log.Debug("debug1")
	log.Warn("warning message")

	rawJSON := []byte(`{
		"level": "info",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "json",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		t.Fatal(err)
	}
	//cfg.DisableStacktrace = false
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	//t.Log(cfg)
	logger, _ := cfg.Build()
	logger.Sugar().Named("rrrrr").Warn("xxxxx")
}

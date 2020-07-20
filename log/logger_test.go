package log

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/qlcchain/qlc-hub/config"
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

func TestInit(t *testing.T) {
	cfg, err := config.DefaultConfig(config.DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}

	err = Setup(cfg)
	if err != nil {
		t.Fatal(err)
	}

	logger := NewLogger("test2")
	logger.Warn("xxxxxxxxxxxxxxxxxxxxxx")
}

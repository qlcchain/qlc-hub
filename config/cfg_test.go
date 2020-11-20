package config

import (
	"os"
	"testing"

	"github.com/jessevdk/go-flags"
)

func TestConfig_Verify(t *testing.T) {
	cfg := &Config{}
	_, _ = flags.ParseArgs(cfg, os.Args)

	type fields struct {
		Verbose  bool
		LogLevel string
		NEOCfg   *NEOCfg
		EthCfg   *EthCfg
		RPCCfg   *RPCCfg
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty fields without default value",
			fields: fields{
				Verbose:  cfg.Verbose,
				LogLevel: cfg.LogLevel,
				NEOCfg:   cfg.NEOCfg,
				EthCfg:   cfg.EthCfg,
				RPCCfg:   cfg.RPCCfg,
			},
			wantErr: true,
		}, {
			name: "empty NEO config",
			fields: fields{
				Verbose:  cfg.Verbose,
				LogLevel: cfg.LogLevel,
				NEOCfg:   nil,
				EthCfg:   cfg.EthCfg,
				RPCCfg:   cfg.RPCCfg,
			},
			wantErr: true,
		}, {
			name: "empty Ethereum config",
			fields: fields{
				Verbose:  cfg.Verbose,
				LogLevel: cfg.LogLevel,
				NEOCfg:   cfg.NEOCfg,
				EthCfg:   nil,
				RPCCfg:   cfg.RPCCfg,
			},
			wantErr: true,
		}, {
			name: "empty RPC config",
			fields: fields{
				Verbose:  cfg.Verbose,
				LogLevel: cfg.LogLevel,
				NEOCfg:   cfg.NEOCfg,
				EthCfg:   cfg.EthCfg,
				RPCCfg:   nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Verbose:  tt.fields.Verbose,
				LogLevel: tt.fields.LogLevel,
				NEOCfg:   tt.fields.NEOCfg,
				EthCfg:   tt.fields.EthCfg,
				RPCCfg:   tt.fields.RPCCfg,
			}
			if err := c.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

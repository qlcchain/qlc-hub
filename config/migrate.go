package config

type CfgMigrate interface {
	Migration(cfg []byte, version int) ([]byte, int, error)
	StartVersion() int
	EndVersion() int
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/qlcchain/qlc-hub/common/util"
)

type CfgManager struct {
	ConfigFile string
	cfgPath    string
	cfg        *Config
}

func NewCfgManager(path string) *CfgManager {
	file := filepath.Join(path, CfgFileName)
	cm := &CfgManager{
		ConfigFile: file,
		cfgPath:    path,
	}
	_, _ = cm.Load()
	return cm
}

func NewCfgManagerWithFile(cfgFile string) *CfgManager {
	return NewCfgManagerWithName(filepath.Dir(cfgFile), filepath.Base(cfgFile))
}

func NewCfgManagerWithName(path string, name string) *CfgManager {
	file := filepath.Join(path, name)
	cm := &CfgManager{
		ConfigFile: file,
		cfgPath:    path,
	}
	_, _ = cm.Load()
	return cm
}

//Load the config file and will create default if config file no exist
// Load the config file and will create default if config file no exist
func (cm *CfgManager) Load(migrations ...CfgMigrate) (*Config, error) {
	_, err := os.Stat(cm.ConfigFile)
	if err != nil {
		err := cm.createAndSave()
		if err != nil {
			return nil, err
		}
	}
	content, err := ioutil.ReadFile(cm.ConfigFile)
	if err != nil {
		return nil, err
	}

	//version, err := cm.parseVersion(content)
	//if err != nil {
	//	fmt.Printf("parse config Version error : %s\n", err)
	//	version = configVersion
	//	err := cm.createAndSave()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	//sort.Slice(migrations, func(i, j int) bool {
	//	if migrations[i].StartVersion() < migrations[j].StartVersion() {
	//		return true
	//	}
	//
	//	if migrations[i].StartVersion() > migrations[j].StartVersion() {
	//		return false
	//	}
	//
	//	return migrations[i].EndVersion() < migrations[j].EndVersion()
	//})
	//for _, m := range migrations {
	//	var err error
	//	if version == m.StartVersion() {
	//		fmt.Printf("migration cfg from v%d to v%d\n", m.StartVersion(), m.EndVersion())
	//		content, version, err = m.Migration(content, version)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//	}
	//}

	// unmarshal as latest config
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	cm.cfg = &cfg
	return &cfg, nil
}

func (cm *CfgManager) createAndSave() error {
	cfg, err := DefaultConfig(cm.cfgPath)
	if err != nil {
		return err
	}
	err = cm.save(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (c *CfgManager) save(cfg interface{}) error {
	dir := filepath.Dir(c.ConfigFile)
	err := util.CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}

	s := util.ToIndentString(cfg)
	return ioutil.WriteFile(c.ConfigFile, []byte(s), 0600)
}

func (c *CfgManager) parseVersion(data []byte) (int, error) {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(data, &objMap)
	if err != nil {
		return 0, err
	}

	if v, ok := objMap["version"]; ok {
		var version int
		if err := json.Unmarshal([]byte(*v), &version); err == nil {
			return version, nil
		} else {
			return 0, err
		}
	} else {
		return 0, errors.New("can not find any version")
	}
}

// Save write config to file
func (cm *CfgManager) Save(data ...interface{}) error {
	dir := filepath.Dir(cm.cfgPath)
	err := util.CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		cfg, err := cm.Load()
		if err != nil {
			return err
		}
		s := util.ToIndentString(cfg)
		return ioutil.WriteFile(cm.ConfigFile, []byte(s), 0600)
	}

	s := util.ToIndentString(data[0])
	return ioutil.WriteFile(cm.ConfigFile, []byte(s), 0600)
}

// Config get current used config
func (cm *CfgManager) Config() (*Config, error) {
	if cm.cfg != nil {
		return cm.cfg, nil
	} else {
		return nil, fmt.Errorf("invalid cfg ,cfg path is [%s]", cm.cfgPath)
	}
}

// ParseDataDir parse dataDir from config file
func (cm *CfgManager) ParseDataDir() (string, error) {
	_, err := os.Stat(cm.ConfigFile)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(cm.ConfigFile)
	if err != nil {
		return "", err
	}

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(content, &objMap)
	if err != nil {
		return "", err
	}

	if v, ok := objMap["dataDir"]; ok {
		var dataDir string
		if err := json.Unmarshal(*v, &dataDir); err == nil {
			return dataDir, nil
		} else {
			return "", err
		}
	} else {
		return "", errors.New("can not parse dataDir")
	}
}

package santaconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"moroz/santa"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

func NewFileRepo(path string) *FileRepo {
	repo := FileRepo{
		configIndex: make(map[string]santa.Config),
		configPath:  path,
	}
	return &repo
}

type FileRepo struct {
	mtx         sync.RWMutex
	configIndex map[string]santa.Config
	configPath  string
}

func (f *FileRepo) updateIndex(configs []santa.Config) {
	f.configIndex = make(map[string]santa.Config, len(configs))
	for _, conf := range configs {
		f.configIndex[conf.MachineID] = conf
	}
}

func (f *FileRepo) AllConfigs(ctx context.Context) ([]santa.Config, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	configs, err := loadConfigs(f.configPath)
	if err != nil {
		return nil, err
	}
	f.updateIndex(configs)
	return configs, nil
}

func (f *FileRepo) Config(ctx context.Context, machineID string) (santa.Config, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	var conf santa.Config
	var class string
	configs, err := loadConfigs(f.configPath)
	if err != nil {
		return conf, errors.Wrapf(err, "loading config for machineID %q", machineID)
	}
	deviceMap, err := loadDeviceMaps(f.configPath)
	if err != nil {
		return conf, errors.Wrapf(err, "loading machineID to class map %q", machineID)
	}
	f.updateIndex(configs)
	if deviceMap[machineID] != "" {
		class = deviceMap[machineID]
	} else {
		class = machineID
	}
	conf, ok := f.configIndex[class]
	if !ok {
		return conf, errors.Errorf("configuration %q not found", machineID)
	}
	return conf, nil
}

func loadConfigs(path string) ([]santa.Config, error) {
	var configs []santa.Config

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(info.Name()) {
		case ".toml":
			file, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				var conf santa.Config
				err := toml.Unmarshal(file, &conf)
				if err != nil {
					return errors.Wrapf(err, "failed to decode %v, skipping \n", info.Name())
				}
				name := info.Name()
				conf.MachineID = strings.TrimSuffix(name, filepath.Ext(name))
				configs = append(configs, conf)
				return nil
			}
		}
		return nil
	})
	return configs, errors.Wrapf(err, "loading configs from path")
}

func loadDeviceMaps(path string) (santa.DeviceMap, error) {
	deviceList := make(santa.DeviceMap)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(info.Name()) {
		case ".csv":
			//class lists
			file, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				name := info.Name()
				class := strings.TrimSuffix(name, filepath.Ext(name))

				devices := strings.Fields(string(file))
				for _, word := range devices {
					deviceList[word] = class
				}
				return nil
			}
		}
		return nil
	})
	return deviceList, errors.Wrapf(err, "loading device maps from path")
}

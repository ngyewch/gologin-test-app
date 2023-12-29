package cmd

import (
	"bytes"
	"encoding/gob"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"path/filepath"
	"strings"
)

func mergeConfig(k *koanf.Koanf, configFile string) error {
	if configFile != "" {
		ext := filepath.Ext(configFile)
		var parser koanf.Parser
		switch ext {
		case ".json":
			parser = json.Parser()
		case ".toml":
			parser = toml.Parser()
		case ".yml":
			parser = yaml.Parser()
		case ".yaml":
			parser = yaml.Parser()
		}
		err := k.Load(file.Provider(configFile), parser)
		if err != nil {
			return err
		}
	}

	err := k.Load(env.Provider("GOLOGIN_", ".", func(s string) string {
		s2 := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "GOLOGIN_")), "_", ".", -1)
		return s2
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func deepCopy(src interface{}, dst interface{}) error {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(src)
	if err != nil {
		return err
	}
	return gob.NewDecoder(&buf).Decode(dst)
}

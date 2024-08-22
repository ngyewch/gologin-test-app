package main

import (
	"github.com/knadh/koanf/v2"
	"github.com/ngyewch/gologin-test-app/server"
	"github.com/urfave/cli/v2"
)

func doServe(cCtx *cli.Context) error {
	configFile := flagConfigFile.Get(cCtx)

	k := koanf.New(".")
	err := mergeConfig(k, configFile)
	if err != nil {
		return err
	}

	var config server.Config
	err = deepCopy(&server.DefaultConfig, &config)
	if err != nil {
		return err
	}

	err = k.Unmarshal("", &config)
	if err != nil {
		return err
	}

	s, err := server.New(&config)
	if err != nil {
		return err
	}
	err = s.Serve()
	if err != nil {
		return err
	}

	return nil
}

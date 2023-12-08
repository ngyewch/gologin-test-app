package cmd

import (
	"fmt"
	"github.com/knadh/koanf/v2"
	"github.com/ngyewch/gologin-test-app/server"
	"github.com/spf13/cobra"
)

var (
	serveCmd = &cobra.Command{
		Use:   fmt.Sprintf("serve [flags]"),
		Short: "Serve",
		Args:  cobra.ExactArgs(0),
		RunE:  serve,
	}
)

func serve(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	k := koanf.New(".")
	err = mergeConfig(k, configFile)
	if err != nil {
		return err
	}

	var config server.Config
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

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("config-file", "", "config file")
}

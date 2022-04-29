package commands

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"git.sr.ht/~glorifiedgluer/mata/config"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	run := func(cmd *cobra.Command, args []string) {
		_ = cmd.Context()

		filePath, err := xdg.ConfigFile(config.ConfigPath)
		if err != nil {
			log.WithField("cmd", cmd.Use).Fatal(err)
		}

		if _, err := os.Stat(filePath); err == nil {
			log.WithField("cmd", cmd.Use).Fatal("config.json already exists")
		} else if errors.Is(err, os.ErrNotExist) {
			body, err := json.MarshalIndent(config.Config{
				Endpoint: "https://mataroa.blog/api",
				Key:      "your-api-key-here",
			}, "", "  ")
			if err != nil {
				log.WithField("cmd", cmd.Use).Fatal(err)
			}

			err = ioutil.WriteFile(filePath, body, os.FileMode((0600)))
			if err != nil {
				log.WithField("cmd", cmd.Use).Fatal(err)
			}

			log.Infof("%s: mata initialized successfully: '%s' file created\n", cmd.Use, filePath)
		}
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize mata",
		Args:  cobra.ExactArgs(0),
		Run:   run,
	}
	return cmd
}

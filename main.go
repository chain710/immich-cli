package main

import (
	"fmt"
	"github.com/chain710/immich-cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"

	log "github.com/sirupsen/logrus"
)

var logLevel string
var apiURL string
var apiKey string
var cfgFile string

//go:generate oapi-codegen -generate "types,client" -package client -o client/immich.auto_generated.go https://raw.githubusercontent.com/immich-app/immich/d2807b8d6ab1a72f37a662423ddda54f41c742ce/server/immich-openapi-specs.json

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".immich")
		viper.SetConfigType("yaml")
	}

	err := viper.ReadInConfig()
	level, err := log.ParseLevel(logLevel)
	cobra.CheckErr(err)
	log.SetLevel(level)

	if err != nil {
		log.Debugf("Can't read config: %v", err)
	} else {
		log.Debugf("Use viper config: %s", viper.ConfigFileUsed())
	}
}

func main() {
	cobra.OnInitialize(initConfig)

	root := &cobra.Command{
		Use: "immich-cli",
	}

	root.AddCommand(cmd.GetAssetsCmd())

	flagSet := pflag.NewFlagSet("flagInConfig", pflag.ContinueOnError)
	flagSet.StringVarP(&logLevel, "log-level", "L", log.InfoLevel.String(), "log level: debug|info|warning|error")
	flagSet.StringVarP(&apiURL, cmd.ViperKey_API, "a", "", "api address, like: https://immich.example.com/api")
	flagSet.StringVarP(&apiKey, cmd.ViperKey_APIKey, "", "", "api key obtained from immich admin")
	cobra.CheckErr(viper.BindPFlags(flagSet))

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.immich)")
	root.PersistentFlags().AddFlagSet(flagSet)
	cobra.CheckErr(root.MarkPersistentFlagRequired(cmd.ViperKey_API))
	cobra.CheckErr(root.MarkPersistentFlagRequired(cmd.ViperKey_APIKey))
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

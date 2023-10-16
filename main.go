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

func initConfig(bindViperFlags *pflag.FlagSet) {
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

	// set flag's value to bypass required check
	bindViperFlags.VisitAll(func(flag *pflag.Flag) {
		if viper.IsSet(flag.Name) {
			_ = bindViperFlags.Set(flag.Name, viper.GetString(flag.Name))
		}
	})
}

func main() {
	var rootCommand = &cobra.Command{
		Use: "immich-cli",
	}

	bindViperFlags := pflag.NewFlagSet("flagInConfig", pflag.ContinueOnError)
	bindViperFlags.StringVarP(&logLevel, "log-level", "L", log.InfoLevel.String(), "log level: debug|info|warning|error")
	bindViperFlags.StringVarP(&apiURL, cmd.ViperKey_API, "a", "", "api address, like: https://immich.example.com/api")
	bindViperFlags.StringVarP(&apiKey, cmd.ViperKey_APIKey, "", "", "api key obtained from immich admin")
	cobra.CheckErr(viper.BindPFlags(bindViperFlags))
	cobra.OnInitialize(func() {
		initConfig(bindViperFlags)
	})

	// set default out as stdout
	rootCommand.SetOut(os.Stdout)
	// add sub commands
	rootCommand.AddCommand(cmd.GetAssetsCmd())
	persistentFlags := rootCommand.PersistentFlags()
	persistentFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.immich)")
	persistentFlags.AddFlagSet(bindViperFlags)
	cobra.CheckErr(rootCommand.MarkPersistentFlagRequired(cmd.ViperKey_API))
	cobra.CheckErr(rootCommand.MarkPersistentFlagRequired(cmd.ViperKey_APIKey))
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

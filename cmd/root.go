package cmd

import (
	"context"
	"os"

	"github.com/creasty/defaults"
	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/homepage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-testnet-homepage",
	Short: "A homepage for Ethereum networks",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := initCommon()
		p := homepage.NewServer(log, cfg)
		if err := p.Start(context.Background()); err != nil {
			log.WithError(err).Fatal("failed to serve")
		}
	},
}

var (
	cfgFile string
	log     = logrus.New()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
}

func loadConfigFromFile(file string) (*homepage.Config, error) {
	if file == "" {
		file = "config.yaml"
	}

	config := &homepage.Config{}

	if err := defaults.Set(config); err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	type plain homepage.Config

	if err := yaml.Unmarshal(yamlFile, (*plain)(config)); err != nil {
		return nil, err
	}

	return config, nil
}

func initCommon() *homepage.Config {
	log.SetFormatter(&logrus.TextFormatter{})

	log.WithField("cfgFile", cfgFile).Info("loading config")

	config, err := loadConfigFromFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	logLevel, err := logrus.ParseLevel(config.Global.LoggingLevel)
	if err != nil {
		log.WithField("logLevel", config.Global.LoggingLevel).Fatal("invalid logging level")
	}

	log.SetLevel(logLevel)

	return config
}

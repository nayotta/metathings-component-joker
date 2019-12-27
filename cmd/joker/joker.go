package main

import (
	"strings"
	"sync"

	cmd_contrib "github.com/nayotta/metathings/cmd/contrib"
	client_helper "github.com/nayotta/metathings/pkg/common/client"
	constant_helper "github.com/nayotta/metathings/pkg/common/constant"
	component "github.com/nayotta/metathings/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	base_opt         *cmd_contrib.BaseOption
	init_config_once sync.Once
)

var (
	RootCmd = &cobra.Command{
		Use: "joker",
	}
)

func initConfig() {
	init_config_once.Do(func() {
		cfg := base_opt.GetConfig()
		if cfg != "" {
			viper.SetConfigFile(cfg)
			if err := viper.ReadInConfig(); err != nil {
				log.WithError(err).Fatalf("failed to read config")
			}
		}
	})
}

func init() {
	opt := cmd_contrib.CreateBaseOption()
	base_opt = &opt

	cobra.OnInitialize(initConfig)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(component.METATHINGS_COMPONENT_PREFIX)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindEnv("stage")

	flags := RootCmd.PersistentFlags()

	flags.StringVar(base_opt.GetLevelP(), "log-level", "info", "Log level")
	flags.StringVarP(base_opt.GetConfigP(), "config", "c", "", "Config file")
	flags.BoolVarP(base_opt.GetVerboseP(), "verbose", "v", false, "Verbose mode")
	flags.StringVar(base_opt.GetKeyFileP(), "key", "", "Transport Credential Key")
	flags.StringVar(base_opt.GetCertFileP(), "cert", "", "Transport Credential Cert")
	flags.BoolVar(base_opt.GetInsecureP(), "insecure", false, "Do not verify transport credential")
	flags.BoolVar(base_opt.GetPlainTextP(), "plaintext", false, "Transport data without tls")
	flags.StringVar(base_opt.GetTokenP(), "token", "", "Access Token")
	flags.StringVar(base_opt.GetServiceEndpoint(client_helper.DEFAULT_CONFIG).GetAddressP(), "addr", constant_helper.CONSTANT_METATHINGSD_DEFAULT_HOST, "MetaThings Service Address")
}

func main() {
	RootCmd.Execute()
}

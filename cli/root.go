package cli

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"httpProxy/config"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

const (
	success       = 0
	invalidConfig = 1
)

var (
	exitCode  = success
	cfgFile   string
	Verbose   bool
	appConfig = &config.ApplicationConfig{
		ClientConfig: config.ClientConfig{
			EnableProxy:   true,
			EnableSocks5:  true,
			HttpProxyPort: 9080,
			Socks5Port:    9022,
			TransportType: config.WebsocketTransport,
		},
		ServerConfig: config.ServerConfig{
			HttpPort: 9180,
		},
	}
	RootCmd = &cobra.Command{
		Use:   "edge-proxy",
		Long:  `This application creates http/socks proxy at edge and transport TCP data to cloud via websocket`,
		Short: `This application creates http/socks proxy at edge and transport TCP data to cloud via websocket`,
	}
)

func Execute(ctx context.Context) {
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path, accept Environment Variable EDGEPROXY_CONFIG (default is $HOME/.edgeproxy/config.yaml) ")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("config") // name of config file (without extension)
	if cfgFile != "" {            // enable ability to specify config file via flag
		log.Debug("cfgFile: ", cfgFile)
		viper.SetConfigFile(cfgFile)
		configDir := path.Dir(cfgFile)
		if configDir != "." && configDir != dir {
			viper.AddConfigPath(configDir)
		}
	}

	viper.AddConfigPath(dir)
	if runtime.GOOS != "windows" {
		viper.AddConfigPath("/etc/edgeproxy")
	} else {
		viper.AddConfigPath("C:\\Program Files\\EdgeProxy\\")
	}
	viper.AddConfigPath("$HOME/.edgeproxy")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("EDGEPROXY")
	viper.AddConfigPath(".")
	viper.BindPFlags(RootCmd.PersistentFlags())

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Errorf("Error when unmarshalling configuration %v", err)
		os.Exit(1)
	}
}

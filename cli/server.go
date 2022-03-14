package cli

import (
	"edgeproxy/server"
	"edgeproxy/server/authorization"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	serverConfig = appConfig.ServerConfig
	serverCmd    = &cobra.Command{
		Use:   "server",
		Short: "Run EdgeProxy as Forward Server",
		Long:  `Run EdgeProxy as Forward Server`,
		Run: func(cmd *cobra.Command, testSuites []string) {
			var err error

			if Verbose {
				log.SetLevel(log.DebugLevel)
				log.Debug("Verbose mode enabled")
			}

			if err = serverConfig.Validate(); err != nil {
				log.Errorf("invalid Server Parameters %v", err)
				os.Exit(invalidConfig)
			}
			var authorizers []authorization.Authorizer = []authorization.Authorizer{}

			if serverConfig.Auth.CaConfig.TrustedRoot != "" {
				authorizers = append(authorizers, authorization.NewSpireAuthorizer(cmd.Context(), serverConfig.Auth.CaConfig))
				//authorizer = authorization.NewSpireAuthorizer(cmd.Context(), serverConfig.Auth.CaConfig)
			} else {
				authorizers = append(authorizers, authorization.NoAuthorizer())
				//authorizer = authorization.NoAuthorizer()
			}

			if serverConfig.FirewallRules != nil {
				authorizers = append(authorizers, authorization.NewFileAuthorizer(cmd.Context(), serverConfig.FirewallRules))
				//authorizer = authorization.NewFileAuthorizer(cmd.Context(), serverConfig.FirewallRules)
			}

			acl := authorization.NewPolicyEnforer(serverConfig)
			webSocketRelay := server.NewHttpServer(cmd.Context(), authorizers, acl, serverConfig.HttpPort)
			webSocketRelay.Start()

			<-cmd.Context().Done()
			webSocketRelay.Stop()
			os.Exit(exitCode)
		},
	}
)

func init() {
	RootCmd.AddCommand(serverCmd)
	clientCmd.PersistentFlags().IntVar(&serverConfig.HttpPort, "http-port", serverConfig.HttpPort, "Http WebSocket Server Listen Port")
}

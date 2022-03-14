package cli

import (
	"edgeproxy/server"
	"edgeproxy/server/auth"
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
			var authenticate auth.Authenticate
			authenticate = auth.NoopAuthorizer()
			if serverConfig.Auth.CaConfig.TrustedRoot != "" {
				authenticate = auth.NewSpireAuthorizer(cmd.Context(), serverConfig.Auth.CaConfig)
			}
			var authorizer auth.Authorize
			authorizer = auth.NoopAuthorizer()
			if serverConfig.Auth.AclPolicyPath != "" {
				authorizer = auth.NewPolicyEnforcer(serverConfig.Auth.AclPolicyPath)

			}
			webSocketRelay := server.NewHttpServer(cmd.Context(), authenticate, authorizer, serverConfig.HttpPort)
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

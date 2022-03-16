package cli

import (
	"edgeproxy/client/clientauth"
	"edgeproxy/client/proxy"
	"edgeproxy/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	clientConfig = appConfig.ClientConfig
	clientCmd    = &cobra.Command{
		Use:   "client",
		Short: "Run EdgeProxy as Client Proxy on edge",
		Long:  `Run EdgeProxy as Client Proxy on edge`,
		Run: func(cmd *cobra.Command, testSuites []string) {
			var err error
			var dialer proxy.Dialer
			var proxyService []proxy.Proxy
			if Verbose {
				log.SetLevel(log.DebugLevel)
				log.Debug("Verbose mode enabled")
			}
			log.Debug(clientConfig)
			if err = clientConfig.Validate(); err != nil {
				log.Errorf("invalid Client Parameters %v", err)
				os.Exit(invalidConfig)
			}
			authenticator, _ := loadAuthenticator()
			switch clientConfig.TransportType {
			case config.WebsocketMuxTransport:
				dialer, err = proxy.NewMuxWebSocketDialer(cmd.Context(), clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, authenticator)
				if err != nil {
					log.Fatal(err)
				}
				break
			case config.WebsocketTransport:
				dialer, err = proxy.NewNoMuxWebSocketDialer(cmd.Context(), clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, authenticator)
				if err != nil {
					log.Fatal(err)
				}
				break
			case config.TcpTransport:
			case config.WireguardTransport:
			case config.UdpTransport:
			case config.QuickTransport:
				panic("Not implemented yet")
			}
			log.Infof("Selected Dialer %s", clientConfig.TransportType)
			if clientConfig.EnableProxy {
				proxyService = append(proxyService, proxy.NewHttpProxy(cmd.Context(), dialer, clientConfig.HttpProxyPort))
			}

			if clientConfig.EnableSocks5 {
				proxyService = append(proxyService, proxy.NewSocksProxy(cmd.Context(), dialer, clientConfig.Socks5Port))
			}

			if len(clientConfig.TransparentProxyList) > 0 {
				proxyService = append(proxyService, proxy.NewTransparentProxy(cmd.Context(), dialer, clientConfig.TransparentProxyList))
			}

			if len(clientConfig.PortForwardList) > 0 {
				proxyService = append(proxyService, proxy.NewPortForwarding(cmd.Context(), dialer, clientConfig.PortForwardList))
			}

			for _, pr := range proxyService {
				pr.Start()
			}
			<-cmd.Context().Done()
			for _, pr := range proxyService {
				pr.Stop()
			}
			os.Exit(exitCode)
		},
	}
)

func loadAuthenticator() (clientauth.Authenticator, error) {
	log.Println(clientConfig.Auth)
	if (clientConfig.Auth.CaConfig != config.ClientAuthCaConfig{}) {
		authenticator := clientauth.JwtAuthenticator{}
		authenticator.Load(clientConfig.Auth.CaConfig)
		return authenticator, nil
	} else {
		return clientauth.NoopAuthenticator{}, nil
	}
}

func init() {

	RootCmd.AddCommand(clientCmd)
	//HTTP Proxy Configuration
	clientCmd.PersistentFlags().BoolVar(&clientConfig.EnableProxy, "http", clientConfig.EnableProxy, "Enable Http Proxy")
	clientCmd.PersistentFlags().IntVar(&clientConfig.HttpProxyPort, "proxy-port", clientConfig.HttpProxyPort, "Http Proxy Listen Port")

	//Socks5 Proxy Configuration
	clientCmd.PersistentFlags().BoolVar(&clientConfig.EnableSocks5, "socks5", clientConfig.EnableSocks5, "Enable Socks5 Proxy")
	clientCmd.PersistentFlags().IntVar(&clientConfig.Socks5Port, "socks5-port", clientConfig.Socks5Port, "Socks5 Proxy Listen Port")

	//Transport Type Configuration
	clientCmd.PersistentFlags().VarP(&clientConfig.TransportType, "transport", "t", "Transport Type")

	//WebSocket Transport Configuration
	clientCmd.PersistentFlags().StringVarP(&clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, "wssTunnelEndpoint", "w", clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, "WebSocket Tunnel endpoint")
	clientCmd.PersistentFlags().VarP(&clientConfig.TransparentProxyList, "transparent-proxy", "k", "Create a transparent Proxy, expected format `5000#TCP#1.1.1.1:5000`")
	clientCmd.PersistentFlags().VarP(&clientConfig.PortForwardList, "port-forward", "f", "Port forward local port to remote TCP service over WebSocket,expected format `5000#TCP#wss://mytunnelendpoint`")

	// TODO: auth config
}

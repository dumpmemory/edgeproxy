package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"httpProxy/client/proxy"
	"httpProxy/client/tcp"
	"httpProxy/client/websocket"
	"httpProxy/config"
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
			var httpProxyService proxy.HTTPProxy
			if Verbose {
				log.SetLevel(log.DebugLevel)
				log.Debug("Verbose mode enabled")
			}

			if err = clientConfig.Validate(); err != nil {
				log.Errorf("invalid Client Parameters %v", err)
				os.Exit(invalidConfig)
			}
			switch clientConfig.TransportType {
			case config.TcpTransport:
				dialer = tcp.NewTCPDialer()
				break
			case config.WebsocketTransport:
				dialer, err = websocket.NewWebSocketDialer(clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint)
				if err != nil {
					log.Fatal(err)
				}
				break
			case config.WireguardTransport:
			case config.UdpTransport:
			case config.QuickTransport:
				panic("Not implemented yet")
			}
			log.Infof("Selected Dialer %s", clientConfig.TransportType)

			httpProxyService = proxy.NewHttpProxy(cmd.Context(), dialer, clientConfig.HttpProxyPort)
			httpProxyService.Start()
			<-cmd.Context().Done()
			os.Exit(exitCode)
		},
	}
)

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
	clientCmd.PersistentFlags().StringVarP(&clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, "wssTunnelEndpoint", "w", clientConfig.WebSocketTransportConfig.WebSocketTunnelEndpoint, "WebSocket Tunnel Endpoint")

}

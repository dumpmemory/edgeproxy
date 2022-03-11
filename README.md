# edgeproxy
Allow Transport TCP Traffic over different proxy transport protocols in a secured and simply way.
It sends all the data over WebSocket to bypass Firewalls and be able to use WAF services like Cloudflare.

## Architecture
```
End Application --> Proxy|Socks|Transparent(EdgeProxy)---- WebSocket ---> (WAF like Cloudflare) -- Websocket---> (EdgeProxy Server) ---> Destination Service
```


## Client Configuration
Client supports:
- Http Proxy Protocol
- Socks5 Proxy Protocol
- Transparent Proxy: Just normal **IP:PORT** binding

### Normal Proxy Client
```
#By default Socks5 and Proxy Server are enabled by default
edgeproxy client --wssTunnelEndpoint https://server.endpoint:9180

#Run curl with proxy configured
curl -x localhost:9080 ifconfig.me
```

### Transparent Proxy
Multiple Proxy Mappings can be provided with `-k`
```
edgeproxy client --wssTunnelEndpoint https://server.endpoint:9180 -k 2222:TCP:192.168.1.194:22 -k 5522:TCP:192.168.1.200:22

#Then run your client application without any proxy configured...
#For example in this case we are mapping SSH servers
ssh localhost -p 2222
ssh localhost -p 5522
```
### Client Help
```
Run EdgeProxy as Client Proxy on edge

Usage:
  edge-proxy client [flags]

Flags:
  -h, --help                                           help for client
      --http                                           Enable Http Proxy
      --http-port int                                  Http WebSocket Server Listen Port (default 9180)
  -f, --port-forward 5000#TCP#wss://mytunnelendpoint   Port forward local port to remote TCP service over WebSocket
,expected format 5000#TCP#wss://mytunnelendpoint (default [])
      --proxy-port int                                 Http Proxy Listen Port (default 9080)
      --socks5                                         Enable Socks5 Proxy
      --socks5-port int                                Socks5 Proxy Listen Port (default 9022)
  -k, --transparent-proxy 5000#TCP#1.1.1.1:5000        Create a transparent Proxy, expected format 5000#TCP#1.1.1.1
:5000 (default [])
  -t, --transport TransportType                        Transport Type (default WebSocketTransport)
  -w, --wssTunnelEndpoint string                       WebSocket Tunnel Endpoint

Global Flags:
      --config string   config file path, accept Environment Variable EDGEPROXY_CONFIG (default is $HOME/.edgeproxy
/config.yaml)
  -v, --verbose         verbose output
```

## Server Configuration
Run server in  the Forwarding Network Zone
```
#Just Run  the server, by default listens on :9180
edgeproxy server

#Optional
#If you don't want to expose this service on internet you can use services like cloudflare tunnel for protecting the service.
cloudflared tunnel --hostname tunnel.edgeproxy.com --url ws://localhost:9180
```

```
Run EdgeProxy as Forward Server

Usage:
  edge-proxy server [flags]

Flags:
  -h, --help   help for server

Global Flags:
      --config string   config file path, accept Environment Variable EDGEPROXY_CONFIG (default is $HOME/.edgeproxy
/config.yaml)
  -v, --verbose         verbose output
```

### Firewall Rules
server can be configured in a way that only allows forward an specific range of IPs.
can be configured if using ``--config /my/config.yml```

#### firewall rules Configuration
```
server:
  firewall:
    defaultAllow: false
    rules:
    - ipnet: 192.168.0.0/24
      ports: [80, 443, 22]
      allow: true
    - ipnet: 0.0.0.0/0
      ports: [80, 443]
      allow: false
```
Server allows configuration hot reloading with ```--watch-config```





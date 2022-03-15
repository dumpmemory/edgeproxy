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


### Certificate Based Authentication

The server can be configured to authenticate clients with a [X.509 SPIFFE Verifiable Identity Document](https://github.com/spiffe/spiffe/blob/main/standards/X509-SVID.md#2-spiffe-id) 
by validating both their cert and a JWT assertion generated from that keypair.

On the server, you have a `root_bundle` for a PEM CA chain of trust, and a [trust_domain](https://github.com/spiffe/spiffe/blob/main/standards/SPIFFE-ID.md#21-trust-domain). 
You can also allow/deny list specific [paths](https://github.com/spiffe/spiffe/blob/main/standards/SPIFFE-ID.md#22-path) 

```yaml
server:
  auth:
    acl:
      ip: resources/example_ip_policy.csv
      domain: resources/example_domain_policy.csv
    ca:
      root_bundle: test/ca.pem
      trust_domain: example.com
      paths:
        allowed:
          - "/users/.*"
        denied:
          - "/users/bad-user.*"
```

Clients then have to be configured to use a SPIFFE compatible client certificate and key

```yaml
client:
  socks5: true
  auth:
    ca:
      key: test/client-key.pem
      cert: test/client.pem
```

Check out [cfssl](https://github.com/cloudflare/cfssl) for an easy way to run a CA.

## Access Control
This project utilizes [casbin](https://github.com/casbin/casbin) to provide a flexible access control policy language.  
Use the `acl` parameter to point to a policy CSV.

There are two ACL files for the server, configured with the following properties in your `config.yaml`
- `server.auth.acl.ip`,  for filtering IP addresses (which is all there is to go on in a SOCKS proxy) 
- `server.auth.acl.doman`, for filtering on hostnames (for example, with an HTTP PROXY)


### IP ACL Entries

```
p, <subject/role>, <IP Address or CIDR>, <port>, <tcp/udp>, <allow/deny>
```

### Domain ACL Entries 
(look very similar, but they match on a domain name instead of being able to do IP network matching) 

```
p, <subject/role>, <*-globbable domain name>, <port>, <tcp/udp>, <allow/deny>
```

### Groups

Groups are supported with a `g` line.  Identities can be associated to groups with `*` glob matching.

For example: 

```casbincsv
g, spiffe://example.com/users/test-*, role_test
```

Would add all users in this trust domain whose name starts with `test-` in to the `role_test` group.

### General Rules
- `deny` takes precedence over `allow`.
- `*` matches are allowed for `subjects` or in `port` matches.


###  Example 
Domain policy file
```casbincsv
p, spiffe://example.com/users/*, ifconfig.me, 443, tcp, allow

p, role_giphy, *.giphy.com, 443, tcp, allow
p, role_giphy, giphy.com, 443, tcp, allow
p, role_giphy, s3.amazonaws.com, 443, tcp, allow


g, spiffe://example.com/users/good-user-123, role_giphy
g, spiffe://example.com/users/ok-user-456, role_giphy
```
- All users can use https://ifconfig.me
- We have a role named `role_giphy`, which is allowed to see images on https://giphy.com
- Users `good-user-123` and `ok-user-456` belong to `role_giphy`, thus are allowed to access https://giphy.com 
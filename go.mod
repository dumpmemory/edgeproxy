module edgeproxy

go 1.17

require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/elazarl/goproxy v0.0.0-20220115173737-adb46da277ac
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
)

require github.com/casbin/casbin/v2 v2.42.0

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3
	github.com/prometheus/client_golang v1.12.1
	gopkg.in/yaml.v2 v2.4.0
)

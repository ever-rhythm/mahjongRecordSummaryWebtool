package client

// ServerAddress majsoul server config
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

// majsoul server http and ws config
type majsoulServerConfig struct {
	Host                string
	Version             string
	Force_version       string
	Code                string
	Ws_server           string
	CodeUpdateTimestamp int64 // http renew majsoul version code after interval
	CodeUpdateInterval  int64
}

var MajsoulServerConfig = majsoulServerConfig{
	Host:               "https://game.maj-soul.com",
	Ws_server:          "wss://gateway-hw.maj-soul.com:443/gateway",
	CodeUpdateInterval: 30,
}

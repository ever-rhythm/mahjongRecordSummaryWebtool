package client

// ServerAddress majsoul server config
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

// majsoul server http and ws config
type majsoulServerConfig struct {
	Host          string
	Version       string
	Force_version string
	Code          string
	Region_urls   [10]string
	Ws_server     string
	Ws_servers    []string
	Interval      int `default:"3"`
}

// net 404 , use com
var MajsoulServerConfig = majsoulServerConfig{
	Host: "https://game.maj-soul.com",
}

package client

// ServerAddress majsoul server config
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

var ServerAddressList = []*ServerAddress{
	/*
		{
			ServerAddress:  "https://game.maj-soul.net", // net 404
			GatewayAddress: "wss://gateway-hw.maj-soul.net/gateway",
			GameAddress:    "wss://gateway-hw.maj-soul.com/game-gateway",
		},

	*/
	// ws connect fail 20231008
	{
		ServerAddress:  "https://game.maj-soul.com",
		GatewayAddress: "wss://gateway-sy.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-sy.maj-soul.com/game-gateway",
	},
}

package config

type Config struct {
	Version  int       `json:"version"`
	DataDir  string    `json:"dataDir"`
	LogLevel string    `json:"logLevel"` //info,warn,debug.
	ChainUrl string    `json:"chainUrl"` //chain url.
	RPC      RPCConfig `json:"rpc"`
	SQLCFG   SqlConfig `json:"sqlcfg"`
}

type RPCConfig struct {
	Enable bool `json:"enabled"`

	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"listenAddress"`

	// TCP or UNIX socket address for the gRPC server to listen on
	GRPCListenAddress  string   `json:"gRPCListenAddress"`
	CORSAllowedOrigins []string `json:"httpCors"`
}

type SqlConfig struct {
	Url string `json:"dbuserurl"`

	Uname string `json:"dbusername"`

	Upwd string `json:"dbuserpasswd"`

	DbName string `json:"dbname"`
}

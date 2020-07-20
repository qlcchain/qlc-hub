package config

type Config struct {
	Version  int        `json:"version"`
	DataDir  string     `json:"dataDir"`
	LogLevel string     `json:"logLevel"` //info,warn,debug.
	ChainUrl string     `json:"chainUrl"` //chain url.
	RPC      RPCConfig  `json:"rpc"`
	P2P      *P2PConfig `json:"p2p"`
}

type RPCConfig struct {
	Enable bool `json:"enabled"`

	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"listenAddress"`

	// TCP or UNIX socket address for the gRPC server to listen on
	GRPCListenAddress  string   `json:"gRPCListenAddress"`
	CORSAllowedOrigins []string `json:"httpCors"`
}

type P2PConfig struct {
	BootNodes          []string `json:"bootNode" mapstructure:"bootNode"`
	IsBootNode         bool     `json:"isBootNode"`
	BootNodeHttpServer string   `json:"bootNodeHttpServer"`
	Listen             string   `json:"listen"`
	// if you are bootNode,should fill in the listening ip
	ListeningIp string `json:"listeningIp"`
	//Time in seconds between sync block interval
	SyncInterval int              `json:"syncInterval"`
	Discovery    *DiscoveryConfig `json:"discovery"`
	ID           *IdentityConfig  `json:"identity" mapstructure:"identity"`
}

type DiscoveryConfig struct {
	// Time in seconds between remote discovery rounds
	DiscoveryInterval int `json:"discoveryInterval"`
	//The maximum number of discovered nodes at a time
	Limit       int  `json:"limit"`
	MDNSEnabled bool `json:"mDNSEnabled"`
	// Time in seconds between local discovery rounds
	MDNSInterval int `json:"mDNSInterval"`
}

type IdentityConfig struct {
	PeerID  string `json:"peerId"`
	PrivKey string `json:"privateKey,omitempty" mapstructure:"privateKey"`
}

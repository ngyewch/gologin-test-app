package server

type Config struct {
	ListenAddress string        `koanf:"listenAddress"`
	BaseUrl       string        `koanf:"baseUrl"`
	SessionName   string        `koanf:"sessionName"`
	SessionSecret string        `koanf:"sessionSecret"`
	Github        *GithubConfig `koanf:"github"`
	Oidc          *OidcConfig   `koanf:"oidc"`
	TcpProxies    []string      `koanf:"tcpProxies"`
}

type GithubConfig struct {
	ClientId     string   `koanf:"clientId"`
	ClientSecret string   `koanf:"clientSecret"`
	Scopes       []string `koanf:"scopes"`
}

type OidcConfig struct {
	IssuerUrl    string   `koanf:"issuerUrl"`
	ClientId     string   `koanf:"clientId"`
	ClientSecret string   `koanf:"clientSecret"`
	Scopes       []string `koanf:"scopes"`
}

var (
	DefaultConfig = Config{
		ListenAddress: ":8080",
		BaseUrl:       "http://127.0.0.1:8080/",
		SessionName:   "gologin-test-app",
		SessionSecret: "gologin-test-app-secret",
	}
)

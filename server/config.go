package server

type Config struct {
	ListenAddress string        `koanf:"listenAddress"`
	BaseUrl       string        `koanf:"baseUrl"`
	SessionName   string        `koanf:"sessionName"`
	SessionSecret string        `koanf:"sessionSecret"`
	Github        *GithubConfig `koanf:"github"`
	Oidc          *OidcConfig   `koanf:"oidc"`
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

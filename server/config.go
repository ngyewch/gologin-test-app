package server

type Config struct {
	ListenAddress string        `koanf:"listenAddress"`
	BaseUrl       string        `koanf:"baseUrl"`
	SessionName   string        `koanf:"sessionName"`
	SessionSecret string        `koanf:"sessionSecret"`
	Github        *GithubConfig `koanf:"github"`
	Oauth2        *Oauth2Config `koanf:"oauth2"`
}

type GithubConfig struct {
	ClientId     string `koanf:"clientId"`
	ClientSecret string `koanf:"clientSecret"`
}

type Oauth2Config struct {
	Endpoint     string   `koanf:"endpoint"`
	ClientId     string   `koanf:"clientId"`
	ClientSecret string   `koanf:"clientSecret"`
	Scopes       []string `koanf:"scopes"`
}

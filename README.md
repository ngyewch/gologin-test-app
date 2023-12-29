# gologin-test-app

## Configuration

| Config path         | Environment variable        | Type     | Default                   | Description          |
|---------------------|-----------------------------|----------|---------------------------|----------------------|
| ListenAddress       | GOLOGIN_LISTENADDRESS       | string   | ":8080"                   | Listen address       |
| BaseUrl             | GOLOGIN_BASEURL             | string   | "http://127.0.0.1:8080/"  | Base URL             |
| SessionName         | GOLOGIN_SESSIONNAME         | string   | "gologin-test-app"        | Web session name     |
| SessionSecret       | GOLOGIN_SESSIONSECRET       | string   | "gologin-test-app-secret" | Web session secret   |
|                     |                             |          |                           |                      |
| Github.ClientId     | GOLOGIN_GITHUB_CLIENTID     | string   |                           | GitHub client ID     |
| Github.ClientSecret | GOLOGIN_GITHUB_CLIENTSECRET | string   |                           | GitHub client secret |
| Github.Scopes       | GOLOGIN_GITHUB_SCOPES       | []string |                           | GitHub scopes        |
|                     |                             |          |                           |                      |
| Oidc.IssuerUrl      | GOLOGIN_OIDC_ISSUERURL      | string   |                           | OIDC issuer URL      |
| Oidc.ClientId       | GOLOGIN_OIDC_CLIENTID       | string   |                           | OIDC client ID       |
| Oidc.ClientSecret   | GOLOGIN_OIDC_CLIENTSECRET   | string   |                           | OIDC client secret   |
| Oidc.Scopes         | GOLOGIN_OIDC_SCOPES         | []string |                           | OIDC scopes          |

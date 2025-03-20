# gologin-test-app

## Configuration

| Config path         | Environment variable        | Type     | Default                   | Description                     |
|---------------------|-----------------------------|----------|---------------------------|---------------------------------|
| listenAddress       | GOLOGIN_LISTENADDRESS       | string   | ":8080"                   | Listen address                  |
| baseUrl             | GOLOGIN_BASEURL             | string   | "http://127.0.0.1:8080/"  | Base URL                        |
| sessionName         | GOLOGIN_SESSIONNAME         | string   | "gologin-test-app"        | Web session name                |
| sessionSecret       | GOLOGIN_SESSIONSECRET       | string   | "gologin-test-app-secret" | Web session secret              |
| tcpProxies          | GOLOGIN_TCPPROXIES          | []string |                           | TCP proxies (for Docker setups) |
|                     |                             |          |                           |                                 |
| github.clientId     | GOLOGIN_GITHUB_CLIENTID     | string   |                           | GitHub client ID                |
| github.clientSecret | GOLOGIN_GITHUB_CLIENTSECRET | string   |                           | GitHub client secret            |
| github.scopes       | GOLOGIN_GITHUB_SCOPES       | []string |                           | GitHub scopes                   |
|                     |                             |          |                           |                                 |
| oidc.issuerUrl      | GOLOGIN_OIDC_ISSUERURL      | string   |                           | OIDC issuer URL                 |
| oidc.clientId       | GOLOGIN_OIDC_CLIENTID       | string   |                           | OIDC client ID                  |
| oidc.clientSecret   | GOLOGIN_OIDC_CLIENTSECRET   | string   |                           | OIDC client secret              |
| oidc.scopes         | GOLOGIN_OIDC_SCOPES         | []string |                           | OIDC scopes                     |

### `tcpProxies`

Format: `<local port>:<remote host>:<remote port>`

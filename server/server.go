package server

import (
	"bytes"
	"encoding/json"
	"github.com/dghubble/gologin/v2"
	gologinGithub "github.com/dghubble/gologin/v2/github"
	gologinOauth2 "github.com/dghubble/gologin/v2/oauth2"
	"github.com/dghubble/sessions"
	"github.com/google/go-github/v48/github"
	"github.com/ngyewch/gologin-test-app/resources"
	"golang.org/x/oauth2"
	oauth2Github "golang.org/x/oauth2/github"
	"html/template"
	"net/http"
	"net/url"
)

const (
	sessionProfile = "profile"
)

type Server struct {
	config       *Config
	sessionStore sessions.Store[any]
	templates    *template.Template
	serveMux     *http.ServeMux
}

type ProfileData struct {
	Type   string        `json:"type"`
	Github *github.User  `json:"github"`
	Oauth2 *oauth2.Token `json:"oauth2"`
}

type LoginTemplateData struct {
	GithubEnabled bool
}

func New(config *Config) (*Server, error) {
	baseUrl, err := url.Parse(config.BaseUrl)
	if err != nil {
		return nil, err
	}
	isSecure := baseUrl.Scheme == "https"

	cookieConfig := sessions.DebugCookieConfig
	if isSecure {
		cookieConfig = sessions.DefaultCookieConfig
	}
	sessionStore := sessions.NewCookieStore[any](cookieConfig, []byte(config.SessionSecret), nil)

	templates, err := template.ParseFS(resources.TemplateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	server := &Server{
		config:       config,
		sessionStore: sessionStore,
		templates:    templates,
	}

	stateConfig := gologin.DebugOnlyCookieConfig
	if isSecure {
		stateConfig = gologin.DefaultCookieConfig
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.profileHandler)
	mux.HandleFunc("/oauth2/logout", server.logoutHandler)

	if config.Github != nil {
		redirectUrl, err := resolveUrl(baseUrl, "/oauth2/github/callback")
		if err != nil {
			return nil, err
		}

		oauth2Config := &oauth2.Config{
			ClientID:     config.Github.ClientId,
			ClientSecret: config.Github.ClientSecret,
			RedirectURL:  redirectUrl.String(),
			Endpoint:     oauth2Github.Endpoint,
		}
		mux.Handle("/oauth2/github/login", gologinGithub.StateHandler(stateConfig, gologinGithub.LoginHandler(oauth2Config, nil)))
		mux.Handle("/oauth2/github/callback", gologinGithub.StateHandler(stateConfig, gologinGithub.CallbackHandler(oauth2Config, http.HandlerFunc(server.issueGithubSession), nil)))
	}

	if config.Oauth2 != nil {
		redirectUrl, err := resolveUrl(baseUrl, "/oauth2/callback")
		if err != nil {
			return nil, err
		}

		endpointUrl, err := url.Parse(config.Oauth2.Endpoint)
		if err != nil {
			return nil, err
		}

		authUrl, err := resolveUrl(endpointUrl, "oauth2/auth")
		if err != nil {
			return nil, err
		}

		tokenUrl, err := resolveUrl(endpointUrl, "oauth2/token")
		if err != nil {
			return nil, err
		}

		endpoint := oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		}
		oauth2Config := &oauth2.Config{
			ClientID:     config.Oauth2.ClientId,
			ClientSecret: config.Oauth2.ClientSecret,
			RedirectURL:  redirectUrl.String(),
			Endpoint:     endpoint,
			Scopes:       config.Oauth2.Scopes,
		}
		mux.Handle("/oauth2/login", gologinOauth2.StateHandler(stateConfig, gologinOauth2.LoginHandler(oauth2Config, nil)))
		mux.Handle("/oauth2/callback", gologinOauth2.StateHandler(stateConfig, gologinOauth2.CallbackHandler(oauth2Config, http.HandlerFunc(server.issueOauth2Session), nil)))
	}

	server.serveMux = mux

	return server, nil
}

func (server *Server) Serve() error {
	return http.ListenAndServe(server.config.ListenAddress, server.serveMux)
}

func (server *Server) saveProfile(w http.ResponseWriter, profile *ProfileData) error {
	jsonBytes, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	session := server.sessionStore.New(server.config.SessionName)
	session.Set(sessionProfile, jsonBytes)
	err = session.Save(w)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) issueGithubSession(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	githubUser, err := gologinGithub.UserFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = server.saveProfile(w, &ProfileData{
		Type:   "github",
		Github: githubUser,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (server *Server) issueOauth2Session(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	token, err := gologinOauth2.TokenFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = server.saveProfile(w, &ProfileData{
		Type:   "github",
		Oauth2: token,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (server *Server) serveLogin(w http.ResponseWriter, req *http.Request) {
	loginTemplateData := LoginTemplateData{
		GithubEnabled: server.config.Github != nil,
	}
	server.serveTemplate(w, req, "login.html", &loginTemplateData)
}

func (server *Server) profileHandler(w http.ResponseWriter, req *http.Request) {
	session, err := server.sessionStore.Get(req, server.config.SessionName)
	if err != nil {
		server.serveLogin(w, req)
		return
	}

	profileBytes, ok := session.Get(sessionProfile).([]byte)
	if !ok {
		server.serveLogin(w, req)
		return
	}

	var profile ProfileData
	err = json.Unmarshal(profileBytes, &profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	server.serveTemplate(w, req, "index.html", &profile)
}

func (server *Server) logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		server.sessionStore.Destroy(w, server.config.SessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (server *Server) serveTemplate(w http.ResponseWriter, req *http.Request, templateName string, data any) {
	buf := bytes.NewBuffer(nil)
	err := server.templates.Lookup(templateName).Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func resolveUrl(baseUrl *url.URL, ref string) (*url.URL, error) {
	refUrl, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return baseUrl.ResolveReference(refUrl), nil
}

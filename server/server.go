package server

import (
	"bytes"
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/dghubble/sessions"
	"github.com/ngyewch/gologin-test-app/resources"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	"html/template"
	"net/http"
	"net/url"
)

const (
	sessionType            = "type"
	sessionGithubUserId    = "githubID"
	sessionGithubUserLogin = "githubUsername"
)

type Server struct {
	config       *Config
	sessionStore sessions.Store[any]
	templates    *template.Template
	serveMux     *http.ServeMux
}

type LoginTemplateData struct {
	GithubEnabled bool
}

type ProfileTemplateData struct {
	Type            string
	GithubUserId    int64
	GithubUserLogin string
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
		callbackUrl, err := url.Parse("/oauth2/github/callback")
		if err != nil {
			return nil, err
		}

		oauth2Config := &oauth2.Config{
			ClientID:     config.Github.ClientId,
			ClientSecret: config.Github.ClientSecret,
			RedirectURL:  baseUrl.ResolveReference(callbackUrl).String(),
			Endpoint:     githubOAuth2.Endpoint,
		}
		mux.Handle("/oauth2/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
		mux.Handle("/oauth2/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, http.HandlerFunc(server.issueGithubSession), nil)))
	}

	server.serveMux = mux

	return server, nil
}

func (server *Server) Serve() error {
	return http.ListenAndServe(server.config.ListenAddress, server.serveMux)
}

func (server *Server) issueGithubSession(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	githubUser, err := github.UserFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session := server.sessionStore.New(server.config.SessionName)
	session.Set(sessionType, "github")
	session.Set(sessionGithubUserId, *githubUser.ID)
	session.Set(sessionGithubUserLogin, *githubUser.Login)
	err = session.Save(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (server *Server) profileHandler(w http.ResponseWriter, req *http.Request) {
	session, err := server.sessionStore.Get(req, server.config.SessionName)
	if err != nil {
		loginTemplateData := LoginTemplateData{
			GithubEnabled: server.config.Github != nil,
		}
		server.serveTemplate(w, req, "login.html", &loginTemplateData)
		return
	}

	sessType := session.Get(sessionType).(string)
	profileTemplateData := ProfileTemplateData{
		Type: sessType,
	}
	switch sessType {
	case "github":
		profileTemplateData.GithubUserId = session.Get(sessionGithubUserId).(int64)
		profileTemplateData.GithubUserLogin = session.Get(sessionGithubUserLogin).(string)
	}

	server.serveTemplate(w, req, "index.html", &profileTemplateData)
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

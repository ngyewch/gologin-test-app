package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dghubble/gologin/v2"
	gologinGithub "github.com/dghubble/gologin/v2/github"
	gologinOauth2 "github.com/dghubble/gologin/v2/oauth2"
	"github.com/dghubble/sessions"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/go-github/v48/github"
	"github.com/hashicorp/cap/oidc"
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
	oidcProvider *oidc.Provider
}

type ProfileData struct {
	Type   string                 `json:"type"`
	Github *github.User           `json:"github"`
	Oidc   *oauth2.Token          `json:"oauth2"`
	Claims map[string]interface{} `json:"claims"`
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

	templates, err := template.ParseFS(resources.TemplateFS, "templates/*.gohtml")
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
			Scopes:       config.Github.Scopes,
		}
		mux.Handle("/oauth2/github/login", gologinGithub.StateHandler(stateConfig, gologinGithub.LoginHandler(oauth2Config, nil)))
		mux.Handle("/oauth2/github/callback", gologinGithub.StateHandler(stateConfig, gologinGithub.CallbackHandler(oauth2Config, http.HandlerFunc(server.issueGithubSession), nil)))
	}

	if config.Oidc != nil {
		redirectUrl, err := resolveUrl(baseUrl, "/oauth2/oidc/callback")
		if err != nil {
			return nil, err
		}

		oidcConfig, err := oidc.NewConfig(config.Oidc.IssuerUrl, config.Oidc.ClientId, oidc.ClientSecret(config.Oidc.ClientSecret), []oidc.Alg{oidc.RS256}, nil)
		if err != nil {
			return nil, err
		}

		oidcProvider, err := oidc.NewProvider(oidcConfig)
		if err != nil {
			return nil, err
		}
		server.oidcProvider = oidcProvider

		discoveryInfo, err := oidcProvider.DiscoveryInfo(context.Background())
		if err != nil {
			return nil, err
		}

		endpoint := oauth2.Endpoint{
			AuthURL:  discoveryInfo.AuthURL,
			TokenURL: discoveryInfo.TokenURL,
		}
		oauth2Config := &oauth2.Config{
			ClientID:     config.Oidc.ClientId,
			ClientSecret: config.Oidc.ClientSecret,
			RedirectURL:  redirectUrl.String(),
			Endpoint:     endpoint,
			Scopes:       config.Oidc.Scopes,
		}
		mux.Handle("/oauth2/oidc/login", gologinOauth2.StateHandler(stateConfig, gologinOauth2.LoginHandler(oauth2Config, nil)))
		mux.Handle("/oauth2/oidc/callback", gologinOauth2.StateHandler(stateConfig, gologinOauth2.CallbackHandler(oauth2Config, http.HandlerFunc(server.issueOidcSession), nil)))
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

func (server *Server) issueOidcSession(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	token, err := gologinOauth2.TokenFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idToken := token.Extra("id_token").(string)
	claims := make(map[string]interface{})
	if idToken != "" {
		fmt.Printf("id_token: %s\n", idToken)
		t, err := jwt.ParseSigned(idToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.UnsafeClaimsWithoutVerification(&claims)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("%v\n", claims)
	}

	err = server.saveProfile(w, &ProfileData{
		Type:   "oidc",
		Oidc:   token,
		Claims: claims,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusFound)
}

func (server *Server) serveLogin(w http.ResponseWriter, req *http.Request) {
	server.serveTemplate(w, req, "login.gohtml", server.config)
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

	server.serveTemplate(w, req, "index.gohtml", &profile)
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

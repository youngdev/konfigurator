package konfigurator

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	oidc "github.com/coreos/go-oidc"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

// OidcGenerator deals with OIDC information such as the redirect endpoint and all the Oauth2 config.
type OidcGenerator struct {
	ctx                   context.Context
	config                oauth2.Config
	localURL              string
	localRedirectEndpoint string
}

// NewOidcGenerator uses a default background context and 'localhost' for the redirectUrl and returns a new OidcGenerator struct.
func NewOidcGenerator(hostURL, clientID, localPort, localRedirectEndpoint string) (*OidcGenerator, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, hostURL)

	if err != nil {
		return nil, err
	}

	localURL := "localhost:" + localPort
	return &OidcGenerator{
		ctx: ctx,
		config: oauth2.Config{
			ClientID:    clientID,
			RedirectURL: "http://" + localURL + localRedirectEndpoint,
			Endpoint:    provider.Endpoint(),
		},
		localURL:              localURL,
		localRedirectEndpoint: localRedirectEndpoint,
	}, nil
}

// AuthCodeURL calls the underlying oauth2.Config AuthCodeURL.
func (o *OidcGenerator) AuthCodeURL(state, nonceValue string) string {
	redirect := url.Values{}
	redirect.Add("client_id", o.config.ClientID)
	redirect.Add("nonce", nonceValue)
	redirect.Add("redirect_uri", o.config.RedirectURL)
	redirect.Add("response_type", "id_token")
	redirect.Add("state", state)
	return fmt.Sprintf("%s?%s", o.config.Endpoint.AuthURL, redirect.Encode())
}

func (o *OidcGenerator) openBrowser() {
	open.Run("http://" + o.localURL)
}

// GetToken retrieves the Oauth2 token from the request and extracts the "id_token" part of it.
func (o *OidcGenerator) GetToken(code string) (string, error) {
	oauth2Token, err := o.config.Exchange(o.ctx, code)
	if err != nil {
		return "", err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", errors.New("missing id_token from oauth2 token")
	}

	return rawIDToken, nil
}

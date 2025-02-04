/*
Copyright 2020 The Flux CD contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitlab

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	gogitlab "github.com/xanzy/go-gitlab"
)

type TokenType string

const (
	// DefaultDomain specifies the default domain used as the backend.
	DefaultDomain = "gitlab.com"

	TokenTypeOAuth2 TokenType = "oauth2"
	TokenTypePat    TokenType = "pat"
	TokenTypeBasic  TokenType = "basicauth"
)

// NewClient creates a new gitlab.Client instance for GitLab API endpoints.
func NewClient(username, password, token string, tokenType TokenType, optFns ...gitprovider.ClientOption) (gitprovider.Client, error) {
	var gl *gogitlab.Client
	var domain, sshDomain string

	// Complete the options struct
	opts, err := gitprovider.MakeClientOptions(optFns...)
	if err != nil {
		return nil, err
	}

	// Create a *http.Client using the transport chain
	httpClient, err := gitprovider.BuildClientFromTransportChain(opts.GetTransportChain())
	if err != nil {
		return nil, err
	}

	if tokenType == TokenTypeOAuth2 {
		if opts.Domain == nil || *opts.Domain == DefaultDomain {
			// No domain set or the default gitlab.com used
			domain = DefaultDomain
			gl, err = gogitlab.NewOAuthClient(token, gogitlab.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
		} else {
			domain = *opts.Domain
			baseURL := fmt.Sprintf("https://%s", domain)
			gl, err = gogitlab.NewOAuthClient(token, gogitlab.WithHTTPClient(httpClient), gogitlab.WithBaseURL(baseURL))
			if err != nil {
				return nil, err
			}
		}
	} else if tokenType == TokenTypePat {
		if opts.Domain == nil || *opts.Domain == DefaultDomain {
			// No domain set or the default gitlab.com used
			domain = DefaultDomain
			gl, err = gogitlab.NewClient(token, gogitlab.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
		} else {
			domain = *opts.Domain
			baseURL := fmt.Sprintf("https://%s", domain)
			gl, err = gogitlab.NewClient(token, gogitlab.WithHTTPClient(httpClient), gogitlab.WithBaseURL(baseURL))
			if err != nil {
				return nil, err
			}
		}
	} else if tokenType == TokenTypeBasic {
		if opts.Domain == nil || *opts.Domain == DefaultDomain {
			// No domain set or the default gitlab.com used
			domain = DefaultDomain
			gl, err = gogitlab.NewBasicAuthClient(username, password, gogitlab.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
		} else {
			domain = *opts.Domain
			baseURL := fmt.Sprintf("https://%s", domain)
			gl, err = gogitlab.NewBasicAuthClient(username, password, gogitlab.WithHTTPClient(httpClient), gogitlab.WithBaseURL(baseURL))
			if err != nil {
				return nil, err
			}
		}
	}

	// By default, turn destructive actions off. But allow overrides.
	destructiveActions := false
	if opts.EnableDestructiveAPICalls != nil {
		destructiveActions = *opts.EnableDestructiveAPICalls
	}

	return newClient(gl, domain, sshDomain, destructiveActions), nil
}

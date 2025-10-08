package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var DebugMode bool

type TokenResponse struct {
	Token struct {
		ID      string `json:"id"`
		Expires string `json:"expires_at"`
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		User struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Domain struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
		} `json:"user"`
		Roles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
		Catalog []struct {
			Type      string `json:"type"`
			ID        string `json:"id"`
			Name      string `json:"name"`
			Endpoints []struct {
				ID        string `json:"id"`
				Interface string `json:"interface"`
				Region    string `json:"region"`
				URL       string `json:"url"`
			} `json:"endpoints"`
		} `json:"catalog"`
	} `json:"token"`
}

// strip / and /v3 from string
func getUrlPath(uri string, suffix string) string {
	trimmed := strings.TrimSuffix(uri, "/")
	host := strings.TrimSuffix(trimmed, "/v3")
	return host + suffix
}

// getDomainSpec returns the domain specification, preferring ID over Name
func getDomainSpec(creds *Credentials) map[string]interface{} {
	if creds.UserDomainId != "" {
		return map[string]interface{}{
			"id": creds.UserDomainId,
		}
	}
	return map[string]interface{}{
		"name": creds.UserDomainName,
	}
}

func GetUnscopedToken(creds *Credentials) (string, error) {
	debugf("GetUnscopedToken called for user %s\n", creds.Username)

	methods := []string{"password"}
	identity := map[string]interface{}{
		"methods": methods,
		"password": map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"password": creds.Password,
			},
		},
	}

	if creds.TOTPCode != "" {
		debugf("Adding TOTP to authentication methods\n")
		methods = append(methods, "totp")
		identity["methods"] = methods
		identity["totp"] = map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"passcode": creds.TOTPCode,
			},
		}
	}

	debugf("Using authentication methods: %v\n", methods)

	authData := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": identity,
		},
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", err
	}

	url := getUrlPath(creds.AuthURL, "/v3/auth/tokens?nocatalog")
	debugf("Making unscoped token request to: %s\n", url)
	debugf("Request body: %s\n", string(jsonData))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debugf("HTTP request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	debugf("HTTP response status: %s (%d)\n", resp.Status, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		debugf("Authentication failed with body: %s\n", string(body))
		return "", fmt.Errorf("authentication failed: %s - %s", resp.Status, string(body))
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		debugf("No X-Subject-Token header received\n")
		return "", fmt.Errorf("no token received")
	}

	debugf("Successfully obtained unscoped token (length: %d)\n", len(token))

	return token, nil
}

func GetApplicationCredentialToken(creds *Credentials) (string, *TokenResponse, error) {
	debugf("GetApplicationCredentialToken called for application credential %s\n", creds.ApplicationCredentialID)

	methods := []string{"application_credential"}
	identity := map[string]interface{}{
		"methods": methods,
		"application_credential": map[string]interface{}{
			"id":     creds.ApplicationCredentialID,
			"secret": creds.ApplicationCredentialSecret,
		},
	}

	authData := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": identity,
		},
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", nil, err
	}

	url := getUrlPath(creds.AuthURL, "/v3/auth/tokens")
	debugf("Making application credential token request to: %s\n", url)
	debugf("Request body: %s\n", string(jsonData))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debugf("HTTP request failed: %v\n", err)
		return "", nil, err
	}
	defer resp.Body.Close()

	debugf("HTTP response status: %s (%d)\n", resp.Status, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		debugf("Application credential authentication failed with body: %s\n", string(body))
		return "", nil, fmt.Errorf("application credential authentication failed: %s - %s", resp.Status, string(body))
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		debugf("No X-Subject-Token header received\n")
		return "", nil, fmt.Errorf("no token received")
	}

	debugf("Successfully obtained application credential token (length: %d)\n", len(token))

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		debugf("Failed to parse token response: %v\n", err)
		return "", nil, fmt.Errorf("failed to parse token response: %v", err)
	}

	debugf("Parsed token response - Project: %s (ID: %s)\n", tokenResponse.Token.Project.Name, tokenResponse.Token.Project.ID)

	return token, &tokenResponse, nil
}

func GetScopedToken(creds *Credentials, projectID string) (string, error) {
	debugf("GetScopedToken called for projectID: %s - always requesting fresh token\n", projectID)

	methods := []string{"password"}
	identity := map[string]interface{}{
		"methods": methods,
		"password": map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"password": creds.Password,
			},
		},
	}

	if creds.TOTPCode != "" {
		debugf("Adding TOTP to scoped authentication (code length: %d)\n", len(creds.TOTPCode))
		methods = append(methods, "totp")
		identity["methods"] = methods
		identity["totp"] = map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"passcode": creds.TOTPCode,
			},
		}
	}

	debugf("Scoped auth using methods: %v\n", methods)

	authData := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": identity,
			"scope": map[string]interface{}{
				"project": map[string]interface{}{
					"id": projectID,
				},
			},
		},
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", err
	}

	url := getUrlPath(creds.AuthURL, "/v3/auth/tokens")
	debugf("Making scoped token request to: %s\n", url)
	debugf("Request body: %s\n", string(jsonData))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debugf("Scoped HTTP request failed: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	debugf("Scoped HTTP response status: %s (%d)\n", resp.Status, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		debugf("Scoped authentication failed with body: %s\n", string(body))
		return "", fmt.Errorf("scoped authentication failed: %s - %s", resp.Status, string(body))
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("no scoped token received")
	}

	debugf("Successfully obtained scoped token (length: %d)\n", len(token))

	return token, nil
}

func GetScopedTokenByProjectName(creds *Credentials, projectName string) (string, *TokenResponse, error) {
	debugf("GetScopedTokenByProjectName called for project: %s\n", projectName)

	methods := []string{"password"}
	identity := map[string]interface{}{
		"methods": methods,
		"password": map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"password": creds.Password,
			},
		},
	}

	if creds.TOTPCode != "" {
		debugf("Adding TOTP to direct scoped authentication (code length: %d)\n", len(creds.TOTPCode))
		methods = append(methods, "totp")
		identity["methods"] = methods
		identity["totp"] = map[string]interface{}{
			"user": map[string]interface{}{
				"name":     creds.Username,
				"domain":   getDomainSpec(creds),
				"passcode": creds.TOTPCode,
			},
		}
	}

	debugf("Direct scoped auth using methods: %v\n", methods)

	scopeData := map[string]interface{}{
		"project": map[string]interface{}{
			"name":   projectName,
			"domain": getDomainSpec(creds),
		},
	}

	authData := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": identity,
			"scope":    scopeData,
		},
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", nil, err
	}

	url := getUrlPath(creds.AuthURL, "/v3/auth/tokens?nocatalog")
	debugf("Making direct scoped token request to: %s\n", url)
	debugf("Request body: %s\n", string(jsonData))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		debugf("Direct scoped HTTP request failed: %v\n", err)
		return "", nil, err
	}
	defer resp.Body.Close()

	debugf("Direct scoped HTTP response status: %s (%d)\n", resp.Status, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		debugf("Direct scoped authentication failed with body: %s\n", string(body))
		return "", nil, fmt.Errorf("scoped authentication failed: %s - %s", resp.Status, string(body))
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		debugf("No X-Subject-Token header received in direct scoped response\n")
		return "", nil, fmt.Errorf("no scoped token received")
	}

	debugf("Successfully obtained direct scoped token (length: %d)\n", len(token))

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		debugf("Failed to parse direct scoped token response: %v\n", err)
		return "", nil, fmt.Errorf("failed to parse token response: %v", err)
	}

	debugf("Parsed token response - Project: %s (ID: %s)\n", tokenResponse.Token.Project.Name, tokenResponse.Token.Project.ID)

	debugf("Direct scoped authentication successful\n")
	return token, &tokenResponse, nil
}

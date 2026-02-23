package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type CredentialFile struct {
	Path        string
	Type        string // "openrc"
	DisplayName string
}

type Credentials struct {
	AuthURL                     string
	Username                    string
	Password                    string
	UserDomainName              string
	UserDomainId                string
	Region                      string
	TOTPCode                    string
	TOTPRequired                bool
	ProjectID                   string
	ProjectName                 string
	SystemScope                 string
	ApplicationCredentialID     string
	ApplicationCredentialSecret string
}

func getPassDir() string {
	passDir := os.Getenv("PASSWORD_STORE_DIR")
	if passDir == "" {
		homeDir, _ := os.UserHomeDir()
		passDir = filepath.Join(homeDir, ".password-store")
	}
	return passDir
}

func GetPassCredFiles() ([]CredentialFile, error) {
	passDir := getPassDir()

	var credFiles []CredentialFile

	err := filepath.Walk(passDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".openrc.gpg") {
			relPath, err := filepath.Rel(passDir, path)
			if err != nil {
				return nil
			}

			passPath := strings.TrimSuffix(relPath, ".gpg")
			passPath = filepath.ToSlash(passPath)

			// Create display name without the extension
			displayName := strings.TrimSuffix(passPath, ".openrc")

			credFiles = append(credFiles, CredentialFile{
				Path:        passPath,
				Type:        "openrc",
				DisplayName: displayName,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(credFiles, func(i, j int) bool {
		return credFiles[i].DisplayName < credFiles[j].DisplayName
	})
	return credFiles, nil
}

// FindCredentialFile searches for a credential file by path or display name
func FindCredentialFile(credFiles []CredentialFile, pathOrName string) CredentialFile {
	// Normalize the input by removing .openrc extension if present
	searchPath := strings.TrimSuffix(pathOrName, ".openrc")

	for _, cf := range credFiles {
		// Match against DisplayName (without .openrc)
		if cf.DisplayName == searchPath {
			return cf
		}
		// Match against Path (with .openrc)
		if cf.Path == pathOrName || cf.Path == searchPath+".openrc" {
			return cf
		}
	}

	return CredentialFile{}
}

func LoadCredentials(credFile CredentialFile) (*Credentials, error) {
	decryptedText, err := passShow(credFile.Path)
	if err != nil {
		return nil, err
	}

	creds := &Credentials{}
	lines := strings.Split(decryptedText, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = after
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "OS_AUTH_URL":
			creds.AuthURL = value
		case "OS_USERNAME":
			creds.Username = value
		case "OS_PASSWORD":
			creds.Password = value
		case "OS_USER_DOMAIN_NAME":
			creds.UserDomainName = value
		case "OS_USER_DOMAIN_ID":
			creds.UserDomainId = value
		case "OS_REGION_NAME":
			creds.Region = value
		case "OS_PROJECT_ID":
			creds.ProjectID = value
		case "OS_PROJECT_NAME":
			creds.ProjectName = value
		case "OS_SYSTEM_SCOPE":
			creds.SystemScope = value
		case "OS_TOTP_REQUIRED":
			creds.TOTPRequired = strings.ToLower(value) == "true" || value == "1"
		case "OS_APPLICATION_CREDENTIAL_ID":
			creds.ApplicationCredentialID = value
		case "OS_APPLICATION_CREDENTIAL_SECRET":
			creds.ApplicationCredentialSecret = value
		}
	}

	// Set default user domain name if neither name nor ID is set
	if creds.UserDomainName == "" && creds.UserDomainId == "" {
		creds.UserDomainName = "Default"
	}

	return creds, nil
}

func passShow(entry string) (string, error) {
	cmd := exec.Command("pass", "show", entry)
	cmd.Env = withPasswordStoreDir(os.Environ(), getPassDir())

	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			msg = ": " + msg
		}
		return "", fmt.Errorf("pass show %q failed: %w%s", entry, err, msg)
	}

	return string(output), nil
}

func withPasswordStoreDir(env []string, passDir string) []string {
	const key = "PASSWORD_STORE_DIR="
	out := make([]string, 0, len(env)+1)
	found := false
	for _, item := range env {
		if strings.HasPrefix(item, key) {
			out = append(out, key+passDir)
			found = true
			continue
		}
		out = append(out, item)
	}
	if !found {
		out = append(out, key+passDir)
	}
	return out
}

// HasProjectDefined returns true if the credentials have a project already specified
func (c *Credentials) HasProjectDefined() bool {
	return c.ProjectID != "" || c.ProjectName != ""
}

// IsApplicationCredential returns true if the credentials use application credential authentication
func (c *Credentials) IsApplicationCredential() bool {
	return c.ApplicationCredentialID != "" && c.ApplicationCredentialSecret != ""
}

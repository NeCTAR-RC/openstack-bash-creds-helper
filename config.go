package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/proglottis/gpgme"
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

func LoadCredentials(credFile CredentialFile) (*Credentials, error) {
	// Initialize GPGME context
	ctx, err := gpgme.New()
	if err != nil {
		return nil, err
	}
	defer ctx.Release()

	// Get password store directory and open encrypted file
	passDir := getPassDir()
	encryptedPath := filepath.Join(passDir, credFile.Path+".gpg")

	file, err := os.Open(encryptedPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create GPGME data from file
	ciphertext, err := gpgme.NewDataFile(file)
	if err != nil {
		return nil, err
	}
	defer ciphertext.Close()

	// Create output buffer for decrypted content
	plaintext, err := gpgme.NewData()
	if err != nil {
		return nil, err
	}
	defer plaintext.Close()

	// Decrypt using GPGME (automatically uses GPG agent)
	err = ctx.Decrypt(ciphertext, plaintext)
	if err != nil {
		return nil, err
	}

	// Read decrypted content
	_, err = plaintext.Seek(0, 0) // Reset to beginning
	if err != nil {
		return nil, err
	}

	// Read all decrypted data
	var decryptedBytes []byte
	buffer := make([]byte, 1024)
	for {
		n, readErr := plaintext.Read(buffer)
		if n > 0 {
			decryptedBytes = append(decryptedBytes, buffer[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	creds := &Credentials{}
	lines := strings.Split(string(decryptedBytes), "\n")

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

// HasProjectDefined returns true if the credentials have a project already specified
func (c *Credentials) HasProjectDefined() bool {
	return c.ProjectID != "" || c.ProjectName != ""
}

// IsApplicationCredential returns true if the credentials use application credential authentication
func (c *Credentials) IsApplicationCredential() bool {
	return c.ApplicationCredentialID != "" && c.ApplicationCredentialSecret != ""
}

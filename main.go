package main

import (
	"flag"
	"fmt"
	"os"
)

var debugMode bool

func debugf(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format, args...)
	}
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	debugMode = *debug
	DebugMode = *debug

	credFiles, err := GetPassCredFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting credential files: %v\n", err)
		os.Exit(1)
	}

	if len(credFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No .openrc files found in pass\n")
		os.Exit(1)
	}

	credFile := SelectCredentialFile(credFiles)
	if credFile.Path == "" {
		fmt.Fprintf(os.Stderr, "No credential file selected\n")
		os.Exit(1)
	}

	// Load credentials from openrc files
	debugf("Loading credentials from %s (type: %s)\n", credFile.Path, credFile.Type)
	creds, err := LoadCredentials(credFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credFile.Path, err)
		os.Exit(1)
	}
	if creds.IsApplicationCredential() {
		debugf("Loaded application credentials - ID: %s, AuthURL: %s\n", creds.ApplicationCredentialID, creds.AuthURL)
	} else {
		debugf("Loaded credentials - Username: %s, AuthURL: %s, TOTPRequired: %v\n", creds.Username, creds.AuthURL, creds.TOTPRequired)
	}
	if creds.ProjectID != "" {
		debugf("ProjectID defined: %s\n", creds.ProjectID)
	}
	if creds.ProjectName != "" {
		debugf("ProjectName defined: %s\n", creds.ProjectName)
	}
	if creds.SystemScope != "" {
		debugf("SystemScope defined: %s\n", creds.SystemScope)
	}

	// Check if TOTP is required and prompt if needed (not applicable for application credentials)
	if !creds.IsApplicationCredential() && creds.TOTPRequired {
		debugf("TOTP required, prompting user\n")
		totpCode, err := PromptForTOTP()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading TOTP code: %v\n", err)
			os.Exit(1)
		}
		creds.TOTPCode = totpCode
		debugf("TOTP code entered (length: %d)\n", len(totpCode))
	} else if creds.IsApplicationCredential() {
		debugf("Application credentials - TOTP not applicable\n")
	} else {
		debugf("TOTP not required\n")
	}

	// If using application credentials, get pre-scoped token directly
	if creds.IsApplicationCredential() {
		debugf("Application credentials detected - getting pre-scoped token\n")

		token, tokenResponse, err := GetApplicationCredentialToken(creds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting application credential token: %v\n", err)
			os.Exit(1)
		}

		debugf("Successfully got application credential token\n")
		selectedProject := &Project{
			ID:   tokenResponse.Token.Project.ID,
			Name: tokenResponse.Token.Project.Name,
		}
		outputEnvironmentVars(credFile, selectedProject, token, creds)
		return
	}

	// If system scope is set, get unscoped token only
	if creds.SystemScope != "" {
		debugf("System scope defined - getting unscoped token only\n")

		token, err := GetUnscopedToken(creds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting unscoped token: %v\n", err)
			os.Exit(1)
		}

		debugf("Successfully got unscoped token for system scope\n")
		outputSystemScopeVars(credFile, token, creds)
		return
	}

	// If project is already defined in the credential file, get scoped token directly
	if creds.HasProjectDefined() {
		debugf("Project defined in credentials - getting scoped token directly\n")

		var scopedToken string
		var selectedProject *Project
		var err error

		if creds.ProjectID != "" {
			debugf("Using ProjectID: %s\n", creds.ProjectID)
			scopedToken, err = GetScopedToken(creds, creds.ProjectID)
			selectedProject = &Project{ID: creds.ProjectID, Name: creds.ProjectName}
		} else {
			debugf("Using ProjectName: %s\n", creds.ProjectName)
			var tokenResponse *TokenResponse
			scopedToken, tokenResponse, err = GetScopedTokenByProjectName(creds, creds.ProjectName)
			if err == nil {
				selectedProject = &Project{
					ID:   tokenResponse.Token.Project.ID,
					Name: tokenResponse.Token.Project.Name,
				}
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting scoped token: %v\n", err)
			os.Exit(1)
		}

		debugf("Successfully got scoped token for project: %s\n", selectedProject.Name)
		outputEnvironmentVars(credFile, selectedProject, scopedToken, creds)
		return
	}

	debugf("No project defined, need to list projects for user selection\n")

	var projectsList []Project

	debugf("Getting unscoped token to list projects\n")
	token, err := GetUnscopedToken(creds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting unscoped token: %v\n", err)
		os.Exit(1)
	}

	debugf("Got unscoped token, listing projects\n")
	projectsList, err = ListProjects(creds.AuthURL, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
		os.Exit(1)
	}
	debugf("Found %d projects\n", len(projectsList))

	if len(projectsList) == 0 {
		fmt.Fprintf(os.Stderr, "No projects found\n")
		os.Exit(1)
	}

	selectedProject := SelectProject(projectsList, credFile)
	if selectedProject == nil {
		fmt.Fprintf(os.Stderr, "No project selected\n")
		os.Exit(1)
	}

	// Update the credential file display name to include the selected project
	credFile.DisplayName = credFile.DisplayName + "/" + selectedProject.Name

	scopedToken, err := GetScopedToken(creds, selectedProject.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting scoped token: %v\n", err)
		os.Exit(1)
	}

	outputEnvironmentVars(credFile, selectedProject, scopedToken, creds)
}

func outputEnvironmentVars(credFile CredentialFile, project *Project, token string, creds *Credentials) {
	fmt.Printf("export OS_CRED=%s\n", credFile.DisplayName)
	fmt.Printf("export OS_IDENTITY_API_VERSION=3\n")
	fmt.Printf("export OS_AUTH_URL=%s\n", creds.AuthURL)
	fmt.Printf("export OS_PROJECT_ID=%s\n", project.ID)
	fmt.Printf("export OS_TOKEN=%s\n", token)
	fmt.Printf("export OS_AUTH_TYPE=token\n")
	if creds.Region != "" {
		fmt.Printf("export OS_REGION_NAME=%s\n", creds.Region)
	}
}

func outputSystemScopeVars(credFile CredentialFile, token string, creds *Credentials) {
	fmt.Printf("export OS_CRED=%s/system\n", credFile.DisplayName)
	fmt.Printf("export OS_IDENTITY_API_VERSION=3\n")
	fmt.Printf("export OS_AUTH_URL=%s\n", creds.AuthURL)
	fmt.Printf("export OS_SYSTEM_SCOPE=%s\n", creds.SystemScope)
	fmt.Printf("export OS_TOKEN=%s\n", token)
	fmt.Printf("export OS_AUTH_TYPE=token\n")
	if creds.Region != "" {
		fmt.Printf("export OS_REGION_NAME=%s\n", creds.Region)
	}
}

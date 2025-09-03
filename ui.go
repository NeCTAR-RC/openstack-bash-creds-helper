package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	fzf "github.com/junegunn/fzf/src"
)

const (
	ColourReset  = "\033[0m"
	ColourRed    = "\033[1;31m"
	ColourYellow = "\033[1;33m"
	ColourPurple = "\033[1;35m"
	ColourGreen  = "\033[1;32m"
)

// keywordColours defines the mapping between environment keywords and their colours
var keywordColours = map[string]string{
	"production/":  ColourRed,
	"rctest/":      ColourYellow,
	"development/": ColourPurple,
}

func removeANSICodes(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func applyKeywordColouring(text string) string {
	result := text
	for keyword, colour := range keywordColours {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return colour + match + ColourReset
		})
	}

	return result
}

func fzfSelect[T any](prompt string, items []T, displayFunc func(T) string) (T, bool) {
	var zero T

	if len(items) == 0 {
		return zero, false
	}

	var itemStrings []string
	var originalTexts []string
	for _, item := range items {
		displayText := displayFunc(item)
		itemStrings = append(itemStrings, displayText)
		originalTexts = append(originalTexts, removeANSICodes(displayText))
	}

	inputChan := make(chan string)
	outputChan := make(chan string)

	// Send items to fzf
	go func() {
		defer close(inputChan)
		for _, item := range itemStrings {
			inputChan <- item
		}
	}()

	var selected string
	var found bool

	// Collect output from fzf
	go func() {
		defer close(outputChan)
		for s := range outputChan {
			selected = s
			found = true
		}
	}()

	// Build fzf options
	options, err := fzf.ParseOptions(
		false, // don't load defaults
		[]string{"--prompt", prompt + " ", "--ansi", "--tac"},
	)
	if err != nil {
		return zero, false
	}

	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	code, _ := fzf.Run(options)
	if code != fzf.ExitOk || !found {
		return zero, false
	}

	selectedClean := removeANSICodes(selected)

	for i, originalText := range originalTexts {
		if originalText == selectedClean {
			return items[i], true
		}
	}

	return zero, false
}

func SelectCredentialFile(credFiles []CredentialFile) CredentialFile {
	if len(credFiles) == 1 {
		return credFiles[0]
	}

	selected, ok := fzfSelect("Select credential file:", credFiles, func(item CredentialFile) string {
		return applyKeywordColouring(item.DisplayName)
	})
	if !ok {
		return CredentialFile{}
	}
	return selected
}

func SelectProject(projectsList []Project, credFile CredentialFile) *Project {
	if len(projectsList) == 1 {
		return &projectsList[0]
	}

	// Determine the colour to use based on the credential file's display name
	cloudColour := getColourForText(credFile.DisplayName)

	selected, ok := fzfSelect("Select project:", projectsList, func(project Project) string {
		// Apply the cloud's colour to the project name
		colouredProjectName := cloudColour + project.Name + ColourReset

		if project.Description != "" {
			return fmt.Sprintf("%s (%s)", colouredProjectName, project.Description)
		}
		return colouredProjectName
	})
	if !ok {
		return nil
	}
	return &selected
}

// getColourForText returns the appropriate colour code for text containing environment keywords
func getColourForText(text string) string {
	lowerText := strings.ToLower(text)
	for keyword, colour := range keywordColours {
		if strings.Contains(lowerText, keyword) {
			return colour
		}
	}
	return "" // No special colour
}

// PromptForTOTP prompts the user to enter a TOTP code
func PromptForTOTP() (string, error) {
	// Open /dev/tty to bypass stderr redirection
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open /dev/tty: %v", err)
	}
	defer tty.Close()

	fmt.Fprint(tty, "Enter TOTP code: ")
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", scanner.Err()
}

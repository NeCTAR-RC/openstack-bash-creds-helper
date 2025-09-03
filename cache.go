package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const CacheExpiryDays = 7

type CacheEntry struct {
	Projects  []Project `json:"projects"`
	Timestamp time.Time `json:"timestamp"`
	AuthURL   string    `json:"auth_url"`
}

type TokenCacheEntry struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CachedAt  time.Time `json:"cached_at"`
}

func getCacheDir() (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(homeDir, ".cache")
	}

	appCacheDir := filepath.Join(cacheDir, "go-creds")
	return appCacheDir, os.MkdirAll(appCacheDir, 0755)
}

func getCacheFilePath(authURL string) (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("projects_%x.json", []byte(authURL))
	return filepath.Join(cacheDir, filename), nil
}

func LoadCachedProjects(authURL string) ([]Project, bool) {
	cacheFile, err := getCacheFilePath(authURL)
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if entry.AuthURL != authURL {
		return nil, false
	}

	if time.Since(entry.Timestamp).Hours() > 24*CacheExpiryDays {
		return nil, false
	}

	return entry.Projects, true
}

func SaveProjectsToCache(authURL string, projectsList []Project) error {
	cacheFile, err := getCacheFilePath(authURL)
	if err != nil {
		return err
	}

	entry := CacheEntry{
		Projects:  projectsList,
		Timestamp: time.Now(),
		AuthURL:   authURL,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func ClearCache(authURL string) error {
	cacheFile, err := getCacheFilePath(authURL)
	if err != nil {
		return err
	}

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(cacheFile)
}

func generateTokenCacheKey(authURL, username, userDomain, projectID string) string {
	data := authURL + "|" + username + "|" + userDomain
	if projectID != "" {
		data += "|" + projectID
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getTokenCacheFilePath(authURL, username, userDomain, projectID string) (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	key := generateTokenCacheKey(authURL, username, userDomain, projectID)
	filename := fmt.Sprintf("token_%s.json", key)
	return filepath.Join(cacheDir, filename), nil
}

func LoadCachedToken(authURL, username, userDomain, projectID string) (string, bool) {
	cacheFile, err := getTokenCacheFilePath(authURL, username, userDomain, projectID)
	if err != nil {
		return "", false
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false
	}

	var entry TokenCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}

	now := time.Now()
	if now.After(entry.ExpiresAt) {
		os.Remove(cacheFile)
		return "", false
	}

	return entry.Token, true
}

func SaveTokenToCache(authURL, username, userDomain, projectID, token, expiresAt string) error {
	cacheFile, err := getTokenCacheFilePath(authURL, username, userDomain, projectID)
	if err != nil {
		return err
	}

	expiryTime, err := time.Parse(time.RFC3339, strings.TrimSuffix(expiresAt, "Z")+"Z")
	if err != nil {
		return err
	}

	entry := TokenCacheEntry{
		Token:     token,
		ExpiresAt: expiryTime,
		CachedAt:  time.Now(),
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func ClearTokenCache(authURL, username, userDomain, projectID string) error {
	cacheFile, err := getTokenCacheFilePath(authURL, username, userDomain, projectID)
	if err != nil {
		return err
	}

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(cacheFile)
}

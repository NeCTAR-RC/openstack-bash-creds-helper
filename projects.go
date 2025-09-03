package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type Project struct {
	ID          string
	Name        string
	Description string
}

type ProjectsResponse struct {
	Projects []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		DomainID    string `json:"domain_id"`
	} `json:"projects"`
}

func ListProjects(authURL, token string) ([]Project, error) {
	authURL = strings.TrimSuffix(authURL, "/")
	req, err := http.NewRequest("GET", authURL+"/v3/auth/projects", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list projects: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var projectsResp ProjectsResponse
	if err := json.Unmarshal(body, &projectsResp); err != nil {
		return nil, err
	}

	var projects []Project
	for _, p := range projectsResp.Projects {
		if p.Enabled {
			projects = append(projects, Project{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
			})
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

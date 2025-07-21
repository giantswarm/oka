package opsgenie

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

// TemplateQuery templates the OpsGenie query string with the provided team and
// the current date.
func TemplateQuery(queryString, team string) (string, error) {
	if queryString == "" {
		return "", fmt.Errorf("query string cannot be empty")
	}

	if team == "" {
		return "", fmt.Errorf("team cannot be empty")
	}

	queryTemplate, err := template.New("opsgenieQuery").Parse(queryString)
	if err != nil {
		return "", fmt.Errorf("failed to parse OpsGenie query template: %w", err)
	}

	queryTemplateData := struct {
		Team  string
		Today string
	}{
		Team:  team,
		Today: time.Now().UTC().Truncate(24 * time.Hour).Format("02-01-2006T15:04:05"),
	}

	var query strings.Builder
	err = queryTemplate.Execute(&query, queryTemplateData)
	if err != nil {
		return "", fmt.Errorf("failed to execute OpsGenie query template: %w", err)
	}

	return query.String(), nil
}

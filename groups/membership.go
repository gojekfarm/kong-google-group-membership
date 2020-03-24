package groups

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// CreateDirectoryService creates the google admin sdk service client instance
func CreateDirectoryService(subject, credentialsPath string) (*admin.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}

	cfg, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryGroupScope)
	if err != nil {
		return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
	}
	cfg.Subject = subject

	ts := cfg.TokenSource(ctx)

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("NewService: %v", err)
	}
	return srv, nil
}

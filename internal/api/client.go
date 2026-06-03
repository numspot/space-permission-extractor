package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"permission-extractor-tui/internal/config"

	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	spaceID    string

	roleCache       map[string]Role
	permissionCache map[string]Permission
	permNameToID    map[string]string
	identityNames   map[string]string
	cacheMu         sync.RWMutex
}

type LogMessage struct {
	Level   string
	Message string
}

func NewClient(cfg config.Config) *Client {
	tokenURL := cfg.TokenURL()

	oauthConfig := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     tokenURL,
	}

	ctx := context.Background()
	httpClient := oauthConfig.Client(ctx)
	httpClient.Timeout = 30 * time.Second

	return &Client{
		httpClient:      httpClient,
		baseURL:         strings.TrimRight(cfg.BaseURL, "/"),
		spaceID:         cfg.SpaceID,
		roleCache:       make(map[string]Role),
		permissionCache: make(map[string]Permission),
		permNameToID:    make(map[string]string),
		identityNames:   make(map[string]string),
	}
}

func (c *Client) fetchAndDecode(ctx context.Context, endpoint string, target interface{}) error {
	fullURL := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Permission-Extractor-TUI/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

func (c *Client) GetUsers(ctx context.Context, logChan chan<- LogMessage) ([]User, error) {
	logChan <- LogMessage{Level: "info", Message: "Fetching users..."}

	base := fmt.Sprintf("/iam/spaces/%s/users", c.spaceID)
	var all []User
	nextToken := ""

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		endpoint := base
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?page[nextToken]=%s", base, url.QueryEscape(nextToken))
		}

		var users struct {
			Items         []User `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}

		if err := c.fetchAndDecode(ctx, endpoint, &users); err != nil {
			return nil, err
		}

		all = append(all, users.Items...)
		logChan <- LogMessage{Level: "info", Message: fmt.Sprintf("Fetched %d users...", len(all))}

		if users.NextPageToken == "" {
			break
		}
		nextToken = users.NextPageToken
	}

	logChan <- LogMessage{Level: "success", Message: fmt.Sprintf("Found %d users", len(all))}
	return all, nil
}

func (c *Client) GetServiceAccounts(ctx context.Context, logChan chan<- LogMessage) ([]ServiceAccount, error) {
	logChan <- LogMessage{Level: "info", Message: "Fetching service accounts..."}

	base := fmt.Sprintf("/iam/spaces/%s/serviceAccounts", c.spaceID)
	var all []ServiceAccount
	nextToken := ""

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		endpoint := base
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?page[nextToken]=%s", base, url.QueryEscape(nextToken))
		}

		var sas struct {
			Items         []ServiceAccount `json:"items"`
			NextPageToken string           `json:"nextPageToken"`
		}

		if err := c.fetchAndDecode(ctx, endpoint, &sas); err != nil {
			return nil, err
		}

		all = append(all, sas.Items...)

		if sas.NextPageToken == "" {
			break
		}
		nextToken = sas.NextPageToken
	}

	logChan <- LogMessage{Level: "success", Message: fmt.Sprintf("Found %d service accounts", len(all))}
	return all, nil
}

func (c *Client) GetUserIamPolicies(ctx context.Context, userID string) (roles []string, permissions []string, err error) {
	endpoint := fmt.Sprintf("/iam/spaces/%s/iampolicy/users/%s", c.spaceID, userID)
	var policy struct {
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	}
	err = c.fetchAndDecode(ctx, endpoint, &policy)
	return policy.Roles, policy.Permissions, err
}

func (c *Client) GetServiceAccountIamPolicies(ctx context.Context, saID string) (roles []string, permissions []string, err error) {
	endpoint := fmt.Sprintf("/iam/spaces/%s/iampolicy/serviceAccounts/%s", c.spaceID, saID)
	var policy struct {
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	}
	err = c.fetchAndDecode(ctx, endpoint, &policy)
	return policy.Roles, policy.Permissions, err
}

func (c *Client) GetUserAcls(ctx context.Context, userID, service, resource string) ([]Acl, error) {
	endpoint := fmt.Sprintf("/iam/spaces/%s/users/%s/acl?service=%s&resource=%s",
		c.spaceID, userID, url.QueryEscape(service), url.QueryEscape(resource))
	var acls struct {
		Acls []Acl `json:"items"`
	}
	err := c.fetchAndDecode(ctx, endpoint, &acls)
	return acls.Acls, err
}

func (c *Client) GetServiceAccountAcls(ctx context.Context, saID, service, resource string) ([]Acl, error) {
	endpoint := fmt.Sprintf("/iam/spaces/%s/serviceAccounts/%s/acl?service=%s&resource=%s",
		c.spaceID, saID, url.QueryEscape(service), url.QueryEscape(resource))
	var acls struct {
		Acls []Acl `json:"items"`
	}
	err := c.fetchAndDecode(ctx, endpoint, &acls)
	return acls.Acls, err
}

func (c *Client) GetRoleDetails(ctx context.Context, roleID string) (Role, error) {
	c.cacheMu.RLock()
	if r, ok := c.roleCache[roleID]; ok {
		c.cacheMu.RUnlock()
		return r, nil
	}
	c.cacheMu.RUnlock()

	endpoint := fmt.Sprintf("/iam/spaces/%s/roles/%s", c.spaceID, roleID)
	var role Role
	if err := c.fetchAndDecode(ctx, endpoint, &role); err != nil {
		return Role{}, err
	}

	c.cacheMu.Lock()
	c.roleCache[roleID] = role
	c.cacheMu.Unlock()

	return role, nil
}

func (c *Client) GetPermissions(ctx context.Context) (map[string]Permission, error) {
	c.cacheMu.RLock()
	if len(c.permissionCache) > 0 {
		result := c.permissionCache
		c.cacheMu.RUnlock()
		return result, nil
	}
	c.cacheMu.RUnlock()

	base := fmt.Sprintf("/iam/spaces/%s/permissions", c.spaceID)
	nextToken := ""

	for {
		endpoint := base
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?page[nextToken]=%s", base, url.QueryEscape(nextToken))
		}

		var list struct {
			Items         []PermissionItem `json:"items"`
			NextPageToken string           `json:"nextPageToken"`
		}

		if err := c.fetchAndDecode(ctx, endpoint, &list); err != nil {
			return nil, err
		}

		c.cacheMu.Lock()
		for _, item := range list.Items {
			c.permissionCache[item.ID] = Permission{
				Name:        item.Name,
				Description: item.Description,
			}
			if item.Name != "" && item.ID != "" {
				c.permNameToID[item.Name] = item.ID
			}
		}
		c.cacheMu.Unlock()

		if list.NextPageToken == "" {
			break
		}
		nextToken = list.NextPageToken
	}

	c.cacheMu.RLock()
	result := c.permissionCache
	c.cacheMu.RUnlock()
	return result, nil
}

func (c *Client) GetPermissionsNameToIDMap(ctx context.Context) (map[string]string, error) {
	c.cacheMu.RLock()
	if len(c.permNameToID) > 0 {
		result := c.permNameToID
		c.cacheMu.RUnlock()
		return result, nil
	}
	c.cacheMu.RUnlock()

	_, err := c.GetPermissions(ctx)
	if err != nil {
		return nil, err
	}

	c.cacheMu.RLock()
	result := c.permNameToID
	c.cacheMu.RUnlock()
	return result, nil
}

func (c *Client) GetPermissionDetails(ctx context.Context, permID string) (Permission, error) {
	c.cacheMu.RLock()
	if p, ok := c.permissionCache[permID]; ok {
		c.cacheMu.RUnlock()
		return p, nil
	}
	c.cacheMu.RUnlock()

	endpoint := fmt.Sprintf("/iam/spaces/%s/permissions/%s", c.spaceID, permID)
	var item PermissionItem
	if err := c.fetchAndDecode(ctx, endpoint, &item); err != nil {
		return Permission{}, err
	}

	p := Permission{Name: item.Name, Description: item.Description}
	c.cacheMu.Lock()
	c.permissionCache[permID] = p
	if item.Name != "" && item.ID != "" {
		c.permNameToID[item.Name] = item.ID
	}
	c.cacheMu.Unlock()

	return p, nil
}

func (c *Client) GetIdentityAccess(ctx context.Context, resourceID string, logChan chan<- LogMessage) ([]AccessRights, error) {
	logChan <- LogMessage{Level: "info", Message: fmt.Sprintf("Fetching identities for resource %s...", resourceID)}

	base := fmt.Sprintf("/iam/spaces/%s/resource/%s", c.spaceID, resourceID)
	var all []AccessRights
	nextToken := ""

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		endpoint := base
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?page[nextToken]=%s", base, url.QueryEscape(nextToken))
		}

		var page struct {
			Items         []AccessRights `json:"items"`
			NextPageToken string         `json:"nextPageToken"`
		}

		if err := c.fetchAndDecode(ctx, endpoint, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Items...)

		if page.NextPageToken == "" {
			break
		}
		nextToken = page.NextPageToken
	}

	logChan <- LogMessage{Level: "success", Message: fmt.Sprintf("Found %d identity access entries", len(all))}
	return all, nil
}

func (c *Client) GetResourcesByIdentity(ctx context.Context, identityID string, logChan chan<- LogMessage) ([]AccessRights, error) {
	logChan <- LogMessage{Level: "info", Message: fmt.Sprintf("Fetching resources for user %s...", identityID)}

	base := fmt.Sprintf("/iam/spaces/%s/users/%s/resources", c.spaceID, identityID)
	var all []AccessRights
	nextToken := ""

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		endpoint := base
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?page[nextToken]=%s", base, url.QueryEscape(nextToken))
		}

		var page struct {
			Items         []AccessRights `json:"items"`
			NextPageToken string         `json:"nextPageToken"`
		}

		if err := c.fetchAndDecode(ctx, endpoint, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Items...)

		if page.NextPageToken == "" {
			break
		}
		nextToken = page.NextPageToken
	}

	logChan <- LogMessage{Level: "success", Message: fmt.Sprintf("Found %d resource entries", len(all))}
	return all, nil
}

func (c *Client) loadIdentityNames(ctx context.Context, logChan chan<- LogMessage) {
	users, err := c.GetUsers(ctx, logChan)
	if err == nil {
		c.cacheMu.Lock()
		for _, u := range users {
			c.identityNames[u.ID] = u.Email
		}
		c.cacheMu.Unlock()
	}

	sas, err := c.GetServiceAccounts(ctx, logChan)
	if err == nil {
		c.cacheMu.Lock()
		for _, sa := range sas {
			c.identityNames[sa.ID] = sa.Name
		}
		c.cacheMu.Unlock()
	}
}

func (c *Client) getIdentityName(uuid string) string {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	return c.identityNames[uuid]
}

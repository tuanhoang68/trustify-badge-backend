package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func AuthURL(shop string, state string) string {
	scopes := os.Getenv("SHOPIFY_SCOPES")
	redirect := os.Getenv("SHOPIFY_APP_URL") + "/api/shopify/callback"
	return fmt.Sprintf(
		"https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		shop,
		os.Getenv("SHOPIFY_API_KEY"),
		url.QueryEscape(scopes),
		url.QueryEscape(redirect),
		state,
	)
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func exchangeToken(shop, code string) (string, error) {
	endpoint := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)
	payload := map[string]string{"client_id": os.Getenv("SHOPIFY_API_KEY"), "client_secret": os.Getenv("SHOPIFY_API_SECRET"), "code": code}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	return tr.AccessToken, nil
}

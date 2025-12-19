package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebhookRegistration struct {
	Topic   string `json:"topic"`
	Address string `json:"address"`
	Format  string `json:"format"`
}

func RegisterWebhook(shop, token, topic, address string) error {
	payload := map[string]map[string]string{
		"webhook": {"topic": topic, "address": address, "format": "json"},
	}
	body, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("https://%s/admin/api/2024-10/webhooks.json", shop)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook reg failed: %d", resp.StatusCode)
	}
	return nil
}

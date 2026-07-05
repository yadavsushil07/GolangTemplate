package razorpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.razorpay.com/v1"

type Client struct {
	keyID     string
	keySecret string
	Order     *OrderResource
}

func NewClient(keyID, keySecret string) *Client {
	c := &Client{keyID: keyID, keySecret: keySecret}
	c.Order = &OrderResource{client: c}
	return c
}

func (c *Client) do(method, path string, body map[string]any) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.keyID, c.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("razorpay error %d: %s", resp.StatusCode, b)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

type OrderResource struct {
	client *Client
}

func (o *OrderResource) Create(data map[string]any, _ map[string]string) (map[string]interface{}, error) {
	return o.client.do("POST", "/orders", data)
}

package dialogapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	endpointSendDialogMessage = "/int/dialog/send"
	endpointGetDialogMessages = "/int/dialog/list"
)

type HTTPAPIClient interface {
	Get(ctx context.Context, path string) (*http.Response, error)
	Post(ctx context.Context, path string, body interface{}) (*http.Response, error)
}

type Client struct {
	apiClient HTTPAPIClient
}

func New(apiClient HTTPAPIClient) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

func (c *Client) SendDialogMessage(ctx context.Context, senderID, receiverID, message string) error {
	response, err := c.apiClient.Post(ctx, endpointSendDialogMessage, map[string]string{
		"from": senderID,
		"to":   receiverID,
		"text": message,
	})
	if err != nil {
		return fmt.Errorf("dialogapiclient failed to send dialog message: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ErrUnexpectedStatusCode
	}

	return nil
}

func (c *Client) GetDialogMessages(ctx context.Context, senderID, receiverID string) ([]DialogMessage, error) {
	response, err := c.apiClient.Get(ctx, fmt.Sprintf("%s?from=%s&to=%s", endpointGetDialogMessages, senderID, receiverID))
	if err != nil {
		return nil, fmt.Errorf("dialogapiclient failed to get dialog messages: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		dialogMessages := make([]DialogMessage, 0)

		err = json.NewDecoder(response.Body).Decode(&dialogMessages)
		if err != nil {
			return nil, fmt.Errorf("dialogapiclient failed to decode api client response: %w", err)
		}

		return dialogMessages, nil
	}

	return nil, ErrUnexpectedStatusCode
}

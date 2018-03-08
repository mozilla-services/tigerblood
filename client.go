package tigerblood

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.mozilla.org/hawk"
)

// Client is an http.Client for the tigerblood service
type Client struct {
	*http.Client
	*hawk.Credentials
	URL string
}

// NewClient creates a new TB client from a base url, hawk ID, and hawk secret
func NewClient(url string, hawkID string, hawkSecret string) (*Client, error) {
	client := &Client{
		Client: &http.Client{},
		Credentials: &hawk.Credentials{
			ID:   hawkID,
			Key:  hawkSecret,
			Hash: sha256.New,
		},
		URL: url,
	}
	return client, nil
}

func (client Client) AuthRequest(req *http.Request, body []byte) {
	req.Header.Set("Content-Type", "application/json")
	auth := hawk.NewRequestAuth(req, client.Credentials, 0)
	hash := auth.PayloadHash("application/json")

	hash.Write(body)
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
}

func (client Client) BanIP(cidr string) (*http.Response, error) {
	entry := ReputationEntry{
		IP:         cidr,
		Reputation: 0,
	}
	body, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Error marshaling request JSON body: %s\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", client.URL, bytes.NewReader(body))
	client.AuthRequest(req, body)
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == http.StatusConflict {
		fmt.Printf("Attempting update of IP since it already exists.\n")

		req, err := http.NewRequest("PUT", client.URL+cidr, bytes.NewReader(body))
		client.AuthRequest(req, body)
		resp, err := client.Do(req)
		if err != nil {
			return resp, err
		}
		if resp.StatusCode != http.StatusOK {
			return resp, errors.New("Unexpected HTTP Status from PUT.")
		}
		return resp, nil
	} else if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Bad response banning IP:\n%+v\n", resp)
		return resp, errors.New("Unexpected HTTP Status from POST.")
	}
	return resp, nil
}

// Exceptions requests the current exceptions list
func (client Client) Exceptions() (*http.Response, error) {
	req, err := http.NewRequest("GET",
		strings.TrimRight(client.URL, "/")+"/exceptions", nil)
	client.AuthRequest(req, []byte{})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad response requesting exceptions:\n%+v\n", resp)
		return resp, errors.New("Unexpected HTTP status from GET.")
	}
	return resp, nil
}

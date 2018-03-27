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

// SetReputation sets the reputation for an IPv4 CIDR to a specific value
func (client Client) SetReputation(cidr string, reputation uint) (*http.Response, error) {
	entry := ReputationEntry{
		IP:         cidr,
		Reputation: reputation,
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
		return resp, errors.New("Unexpected HTTP Status from POST.")
	}
	return resp, nil
}

// BanIP sets the reputation for an IPv4 CIDR to 0 to block it for the maximum decay period
func (client Client) BanIP(cidr string) (*http.Response, error) {
	resp, error := client.SetReputation(cidr, 0)
	if error == errors.New("Unexpected HTTP Status from POST.") {
		fmt.Printf("Bad response banning IP:\n%+v\n", resp)
		return resp, error
	} else if error != nil {
		return resp, error
	}
	return resp, nil
}


// UnbanIP sets the reputation for an IPv4 CIDR to 100 to immediately unblock it
func (client Client) UnbanIP(cidr string) (*http.Response, error) {
	resp, error := client.SetReputation(cidr, 0)
	if error == errors.New("Unexpected HTTP Status from POST.") {
		fmt.Printf("Bad response unbanning IP:\n%+v\n", resp)
	} else if error != nil {
		return resp, error
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

// Reputation requests the reputation score for an IP address
func (client Client) Reputation(ipaddr string) (*http.Response, error) {
	req, err := http.NewRequest("GET",
		strings.TrimRight(client.URL, "/")+"/"+ipaddr, nil)
	client.AuthRequest(req, []byte{})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad response requesting reputation:\n%+v\n", resp)
		return resp, errors.New("Unexpected HTTP status from GET.")
	}
	return resp, nil
}

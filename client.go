package tigerblood

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"go.mozilla.org/hawk"
)

var (
	ClientUnexpectedGETStatusError  = errors.New("Unexpected HTTP status from GET")
	ClientUnexpectedPUTStatusError  = errors.New("Unexpected HTTP Status from PUT")
	ClientUnexpectedPOSTStatusError = errors.New("Unexpected HTTP Status from POST")
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

// SetReputation sets the reputation for an IPv4 CIDR to a specific value. If rev is set to
// true, the reputation entry also has it's reviewed flag set to true in the database.
func (client Client) SetReputation(cidr string, reputation uint, rev bool) (*http.Response, error) {
	entry := ReputationEntry{
		IP:         cidr,
		Reputation: reputation,
		Reviewed:   rev,
	}
	body, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", client.URL+cidr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	client.AuthRequest(req, body)
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, ClientUnexpectedPUTStatusError
	}
	return resp, nil
}

// SetReviewed sets the review flag for a given CIDR to status
func (client Client) SetReviewed(cidr string, status bool) (*http.Response, error) {
	resp, err := client.Reputation(cidr)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r ReputationEntry
	err = json.Unmarshal(buf, &r)
	if err != nil {
		return nil, err
	}

	r.Reviewed = status
	buf, err = json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", client.URL+cidr, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	client.AuthRequest(req, buf)
	resp, err = client.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, ClientUnexpectedPUTStatusError
	}
	return resp, nil
}

// BanIP sets the reputation for an IPv4 CIDR to 0 to block it for the maximum decay period
func (client Client) BanIP(cidr string) (*http.Response, error) {
	// Since this is being applied from the ban command, set reviewed to true
	return client.SetReputation(cidr, 0, true)
}

// UnbanIP sets the reputation for an IPv4 CIDR to 100 to immediately unblock it
func (client Client) UnbanIP(cidr string) (*http.Response, error) {
	return client.SetReputation(cidr, 100, false)
}

// Exceptions requests the current exceptions list
func (client Client) Exceptions() (*http.Response, error) {
	req, err := http.NewRequest("GET",
		strings.TrimRight(client.URL, "/")+"/exceptions", nil)
	if err != nil {
		return nil, err
	}
	client.AuthRequest(req, []byte{})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, ClientUnexpectedGETStatusError
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
		return resp, ClientUnexpectedGETStatusError
	}
	return resp, nil
}

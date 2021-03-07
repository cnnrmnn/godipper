package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/graphql-go/graphql"
)

var twilioAccountSID = os.Getenv("TWILIO_ACCOUNT_SID")
var twilioAuthToken = os.Getenv("TWILIO_AUTH_TOKEN")
var twilioServiceID = os.Getenv("TWILIO_SERVICE_ID")
var twilioURL = "https://verify.twilio.com/v2/Services/" + twilioServiceID

// twilioPost makes an HTTP POST request to the given url with basic
// authentication credentials set. It url encodes the values map as the request
// body.
func twilioPost(url string, body url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating Twilio request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(twilioAccountSID, twilioAuthToken)
	return http.DefaultClient.Do(req)
}

// sendToken sends a verification token to the given phone number. The phone
// number should be a ten-digit US number.
func sendToken(to string) error {
	u := twilioURL + "/Verifications"
	body := url.Values{
		"To":      {"+1" + to},
		"Channel": {"sms"},
	}
	resp, err := twilioPost(u, body)
	if err != nil {
		return fmt.Errorf("sending verification code: %v", err)
	}
	resp.Body.Close()
	if resp.Status != "201 Created" {
		return errors.New("failed to send verification code")
	}
	return nil
}

// checkToken returns true if the given verification token is valid for the
// given phone number.
func checkToken(to, code string) (bool, error) {
	u := twilioURL + "/VerificationCheck"
	body := url.Values{
		"To":   {"+1" + to},
		"Code": {code},
	}
	resp, err := twilioPost(u, body)
	if err != nil {
		return false, fmt.Errorf("posting verification check: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading verification response body: %v", err)
	}

	var decoded interface{}
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		return false, fmt.Errorf("parsing verification response body: %v", err)
	}
	status, ok := decoded.(map[string]interface{})["status"].(string)
	if !ok {
		return false, fmt.Errorf("failed to find status: %v", err)
	}
	return status == "approved", nil
}

func sendCode() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Phone",
				Fields: graphql.Fields{
					"exists": &graphql.Field{Type: graphql.Boolean},
				},
			}),
		Args: graphql.FieldConfigArgument{
			"phone": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			phone := p.Args["phone"].(string)
			if err := sendToken(phone); err != nil {
				return nil, err
			}
			return map[string]interface{}{"exists": false}, nil
		},
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"github.com/jmoiron/jsonq"
	"strings"

	"net/http"
	"net/url"
)

func handleAuth(clientID string, clientSecret string, w http.ResponseWriter, r *http.Request) {
	authCode, ok := r.URL.Query()["code"]

	if !ok || len(authCode[0]) < 1 {
		logger.Error("Url Param 'code' is missing")
		http.Error(w, "You did not provide a code in the url", 400)
		return
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", authCode[0])

	logger.Infof("Going to submit: %v", data)

	resp, err := http.Post("https://slack.com/api/oauth.access", "application/x-www-form-urlencoded; charset=utf-8", strings.NewReader(data.Encode()))

	if err != nil {
		logger.Errorf("There was an error from slack: %v", err)
		http.Error(w, "There was an error authenticating with slack.", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Status error: %v", resp.StatusCode)
	}

	respData := map[string]interface{}{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&respData)
	logger.Infof("Resposne data: %v", respData)
	jq := jsonq.NewQuery(respData)

	if slackOk, err := jq.Bool("ok"); err == nil && slackOk {
		accessToken, err := jq.String("access_token")
		if err == nil {
			redirectToTeam(accessToken, w, r)
		} else {
			fmt.Fprintf(w, "Success installing into Slack!")
			logger.Infof("Success installing into Slack!")
		}
	} else {
		slackError, err := jq.String("error")
		if err != nil {
			slackError = "Failed to parse response"
		}
		fmt.Fprintf(w, "Failed to auth with slack: %s", slackError)
		logger.Errorf("Failed to auth with slack: %s", slackError)
	}
}

func redirectToTeam(accessToken string, w http.ResponseWriter, r *http.Request) {
	data := url.Values{}
	data.Set("token", accessToken)

	logger.Infof("Going to submit: %v", data)

	resp, err := http.Post("https://slack.com/api/team.info", "application/x-www-form-urlencoded; charset=utf-8", strings.NewReader(data.Encode()))

	if err != nil {
		logger.Errorf("There was an error from slack: %v", err)
		http.Error(w, "There was an error authenticating with slack.", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Status error: %v", resp.StatusCode)
	}

	respData := map[string]interface{}{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&respData)
	logger.Infof("Resposne data: %v", respData)
	jq := jsonq.NewQuery(respData)

	if slackOk, err := jq.Bool("ok"); err == nil && slackOk {
		domain, err := jq.String("team", "domain")
		if err == nil {
			logger.Infof("Success installing into Slack %s!", domain)
			http.Redirect(w, r, fmt.Sprintf("https://%s.slack.com", domain), http.StatusSeeOther)
		} else {
			fmt.Fprintf(w, "Success installing into Slack!")
			logger.Infof("Success installing into Slack!")
		}
	} else {
		slackError, err := jq.String("error")
		if err != nil {
			slackError = "Failed to parse response"
		}
		fmt.Fprintf(w, "Failed to get team from slack: %s", slackError)
		logger.Errorf("Failed to get team from slack: %s", slackError)
	}
}

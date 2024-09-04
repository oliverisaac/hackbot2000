package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"encoding/json"
	"github.com/google/logger"
)

var user_regex = regexp.MustCompile(`<@([a-z0-9A-Z]+)([|][^>]+)?>`)
var leader_regex = regexp.MustCompile(`^lead`)
var token string

const logPath = "./hackerbot2000.log"

func handleLeaders(team_id string) (string, string, error) {
	var response string
	for rank, leader := range getLeaders(team_id) {
		score := leader.Score
		if score < -5 {
			score = -5
		}
		thisLine := fmt.Sprintf("%d: <@%s> (%d pts.)", (rank + 1), leader.User, score)
		if response == "" {
			response = thisLine
		} else {
			response = fmt.Sprintf("%s\n%s", response, thisLine)
		}
	}
	return response, "ephemeral", nil
}

func userStringToUserID(user string) string {
	result := user_regex.FindStringSubmatch(user)[1]
	return result
}

func getUserScore(user string, team string) int {
	user_score := getTimesHacker(user, team) - getTimesVictim(user, team)
	if user_score < -5 {
		return -5
	}

	return user_score
}

func handleHack(team_id string, victim string, hacker string) (string, string, error) {
	if victim == hacker {
		return "You can't hack yourself.", "ephemeral", nil
	}

	if recentlyHacked(victim, team_id) {
		return fmt.Sprintf("<@%s> has been too recently hacked.", victim), "ephemeral", nil
	}

	addHack(victim, hacker, team_id)

	hacker_score := getUserScore(hacker, team_id)
	victim_score := getUserScore(victim, team_id)
	msg := fmt.Sprintf("<@%s> (%d pts.) was hacked by <@%s> (%d pts.)!", victim, victim_score, hacker, hacker_score)
	resp_type := "in_channel"

	return msg, resp_type, nil
}

func hackHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var responseType string
	var responseMessage string

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if token != r.FormValue("token") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	text := strings.Replace(r.FormValue("text"), "\r", "", -1)
	logger.Info("Text: " + text)

	from_user := strings.Replace(r.FormValue("user_id"), "\r", "", -1)
	logger.Info("From User: " + from_user)

	from_team := strings.Replace(r.FormValue("team_id"), "\r", "", -1)
	from_enterprise := strings.Replace(r.FormValue("enterprise_id"), "\r", "", -1)
	if from_enterprise != "" {
		from_team = from_enterprise + ":" + from_team
	}
	logger.Info("Team: " + from_team)

	if user_regex.Match([]byte(text)) {
		tagged_user := string(user_regex.Find([]byte(text)))
		tagged_user = userStringToUserID(tagged_user)
		responseMessage, responseType, err = handleHack(from_team, tagged_user, from_user)
	} else if leader_regex.Match([]byte(text)) {
		responseMessage, responseType, err = handleLeaders(from_team)
	} else {
		responseMessage = "Please either tag an unlocked user `/hack @username` or use `/hack leaders`"
		responseType = "ephemeral"
		err = nil
	}

	if err != nil {
		logger.Error(err)
		responseMessage = "There was an error."
		responseType = "ephemeral"
	}

	jsonResp, _ := json.Marshal(struct {
		Type string `json:"response_type"`
		Text string `json:"text"`
	}{
		Type: responseType,
		Text: responseMessage,
	})

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResp))
}

func main() {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("LoggerExample", true, false, lf).Close()

	cfg := doConfig()

	dbConfig := dbConnection{
		host:     cfg.GetString("db.host"),
		port:     cfg.GetInt("db.port"),
		username: cfg.GetString("db.username"),
		password: cfg.GetString("db.password"),
		name:     cfg.GetString("db.name"),
		options:  "charset=utf8&parseTime=True",
	}

	db := dbInit(dbConfig)
	defer db.Close()

	token = cfg.GetString("slack.token")
	logger.Info(fmt.Sprintf("Going to listen on port %d", cfg.GetInt("port")))
	http.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		handleAuth(cfg.GetString("slack.clientID"), cfg.GetString("slack.clientSecret"), w, r)
	})
	http.HandleFunc("/hack", hackHandler)
	logger.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("port")), nil))
}

package utils_twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	loginURL     = "https://api.twitter.com/1.1/onboarding/task.json"
	bearerToken2 = "AAAAAAAAAAAAAAAAAAAAAFQODgEAAAAAVHTp76lzh3rFzcHbmHVvQxYYpTw%3DckAlMINMjmCwxUcaXbAN4XqJVdgMJaHqNOFgPMK0zN1qLqLQCF"
	userAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"
	tokensDir    = "tokens/twitter"
)

var (
	globalClient     *http.Client
	globalGuestToken string
	isLoggedIn       bool
	loginInProgress  = make(map[string]bool)
	loginMutex       sync.RWMutex
)

type flow struct {
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	FlowToken string `json:"flow_token"`
	Status    string `json:"status"`
	Subtasks  []struct {
		SubtaskID string `json:"subtask_id"`
	} `json:"subtasks"`
}

type SessionData struct {
	GuestToken  string                 `json:"guest_token"`
	Cookies     []*http.Cookie         `json:"cookies"`
	BearerToken string                 `json:"bearer_token"`
	CSRFToken   string                 `json:"csrf_token"`
	AuthToken   string                 `json:"auth_token"`
	SessionData map[string]interface{} `json:"session_data"`
	LoginTime   time.Time              `json:"login_time"`
}

func SaveTokensForUser(userID string) error {
	if globalClient == nil {
		return fmt.Errorf("no session to save")
	}

	cookies := []*http.Cookie{}
	var csrfToken, authToken string

	if globalClient.Jar != nil {
		u, _ := url.Parse("https://twitter.com")
		cookies = globalClient.Jar.Cookies(u)

		for _, cookie := range cookies {
			switch cookie.Name {
			case "ct0":
				csrfToken = cookie.Value
			case "auth_token":
				authToken = cookie.Value
			}
		}
	}

	session := SessionData{
		GuestToken:  globalGuestToken,
		Cookies:     cookies,
		BearerToken: bearerToken2,
		CSRFToken:   csrfToken,
		AuthToken:   authToken,
		LoginTime:   time.Now(),
		SessionData: map[string]interface{}{
			"user_agent": userAgent,
			"login_url":  loginURL,
		},
	}

	os.MkdirAll(tokensDir, 0755)
	tokensFile := fmt.Sprintf("%s/tokens-%s.json", tokensDir, userID)
	data, _ := json.MarshalIndent(session, "", "  ")
	return os.WriteFile(tokensFile, data, 0644)
}

func LoadTokensForUser(userID string) error {
	tokensFile := fmt.Sprintf("%s/tokens-%s.json", tokensDir, userID)
	data, err := os.ReadFile(tokensFile)
	if err != nil {
		return err
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	if time.Since(session.LoginTime) > 24*time.Hour {
		return fmt.Errorf("session expired (older than 24 hours)")
	}

	jar, _ := cookiejar.New(nil)
	globalClient = &http.Client{Jar: jar, Timeout: 60 * time.Second}

	if len(session.Cookies) > 0 {
		u, _ := url.Parse("https://twitter.com")
		globalClient.Jar.SetCookies(u, session.Cookies)

		u2, _ := url.Parse("https://x.com")
		globalClient.Jar.SetCookies(u2, session.Cookies)
	}

	globalGuestToken = session.GuestToken
	isLoggedIn = true

	fmt.Printf("Loaded session for user %s (saved %v ago)\n", userID, time.Since(session.LoginTime).Round(time.Minute))
	return nil
}

func ValidateSession() bool {
	req, _ := http.NewRequest("GET", "https://twitter.com/i/api/graphql/ldqoq5MmFHN1FhMGvzC9Jg/TweetDetail", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken2)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Guest-Token", globalGuestToken)

	query := url.Values{}
	query.Set("variables", `{"focalTweetId":"1"}`)
	query.Set("features", `{}`)
	req.URL.RawQuery = query.Encode()

	resp, err := globalClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode != 404
}

func randomDelay() {
	delay := time.Duration(3000+rand.Intn(5000)) * time.Millisecond
	time.Sleep(delay)
}

func getGuestToken(client *http.Client) (string, error) {
	req, _ := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken2)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var jsn map[string]interface{}
	json.Unmarshal(body, &jsn)
	guestToken := jsn["guest_token"].(string)
	return guestToken, nil
}

func getFlowToken(client *http.Client, guestToken string, data map[string]interface{}) (string, error) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", loginURL, bytes.NewReader(jsonData))

	req.Header.Set("Authorization", "Bearer "+bearerToken2)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Guest-Token", guestToken)
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Client")
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Client-Language", "en")

	for _, cookie := range client.Jar.Cookies(req.URL) {
		if cookie.Name == "ct0" {
			req.Header.Set("X-CSRF-Token", cookie.Value)
			break
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info flow
	json.Unmarshal(body, &info)

	if len(info.Errors) > 0 {
		return "", fmt.Errorf("auth error (%d): %v", info.Errors[0].Code, info.Errors[0].Message)
	}

	return info.FlowToken, nil
}

func LoginAndSaveTokens(username, password, userID string) error {
	fmt.Printf("⚠️  WARNING: Using username/password login. This may trigger account restrictions.\n")
	fmt.Printf("Starting Twitter login for user: %s\n", username)

	jar, _ := cookiejar.New(nil)
	globalClient = &http.Client{Jar: jar, Timeout: 60 * time.Second}

	guestToken, err := getGuestToken(globalClient)
	if err != nil {
		return fmt.Errorf("error getting guest token: %v", err)
	}
	globalGuestToken = guestToken

	randomDelay()

	data := map[string]interface{}{
		"flow_name": "login",
		"input_flow_data": map[string]interface{}{
			"flow_context": map[string]interface{}{
				"debug_overrides": map[string]interface{}{},
				"start_location":  map[string]interface{}{"location": "splash_screen"},
			},
		},
	}
	flowToken, err := getFlowToken(globalClient, guestToken, data)
	if err != nil {
		return fmt.Errorf("error in flow start: %v", err)
	}

	randomDelay()

	data = map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id":         "LoginJsInstrumentationSubtask",
				"js_instrumentation": map[string]interface{}{"response": "{}", "link": "next_link"},
			},
		},
	}
	flowToken, err = getFlowToken(globalClient, guestToken, data)
	if err != nil {
		return fmt.Errorf("error in instrumentation: %v", err)
	}

	randomDelay()

	data = map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id": "LoginEnterUserIdentifierSSO",
				"settings_list": map[string]interface{}{
					"setting_responses": []map[string]interface{}{
						{
							"key":           "user_identifier",
							"response_data": map[string]interface{}{"text_data": map[string]interface{}{"result": username}},
						},
					},
					"link": "next_link",
				},
			},
		},
	}
	flowToken, err = getFlowToken(globalClient, guestToken, data)
	if err != nil {
		return fmt.Errorf("error submitting username: %v", err)
	}

	randomDelay()

	data = map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id":     "LoginEnterPassword",
				"enter_password": map[string]interface{}{"password": password, "link": "next_link"},
			},
		},
	}
	flowToken, err = getFlowToken(globalClient, guestToken, data)
	if err != nil {
		return fmt.Errorf("error submitting password: %v", err)
	}

	randomDelay()

	data = map[string]interface{}{
		"flow_token": flowToken,
		"subtask_inputs": []map[string]interface{}{
			{
				"subtask_id":              "AccountDuplicationCheck",
				"check_logged_in_account": map[string]interface{}{"link": "AccountDuplicationCheck_false"},
			},
		},
	}
	_, err = getFlowToken(globalClient, guestToken, data)
	if err != nil {
		if strings.Contains(err.Error(), "LoginAcid") || strings.Contains(err.Error(), "LoginTwoFactorAuthChallenge") {
			return fmt.Errorf("2FA/Email confirmation required: %v", err)
		} else {
			return fmt.Errorf("error in duplication check: %v", err)
		}
	}

	isLoggedIn = true
	fmt.Println("Login successful!")

	if err := SaveTokensForUser(userID); err != nil {
		fmt.Printf("Warning: Could not save tokens: %v\n", err)
	} else {
		fmt.Printf("Session saved for user %s\n", userID)
	}

	return nil
}

func StartLoginAsync(username, password, userID string) {
	loginMutex.Lock()
	if loginInProgress[userID] {
		loginMutex.Unlock()
		return
	}
	loginInProgress[userID] = true
	loginMutex.Unlock()

	go func() {
		defer func() {
			loginMutex.Lock()
			delete(loginInProgress, userID)
			loginMutex.Unlock()
		}()

		err := LoginAndSaveTokens(username, password, userID)
		if err != nil {
			fmt.Printf("Background login failed for user %s: %v\n", userID, err)
		}
	}()
}

func IsLoginInProgress(userID string) bool {
	loginMutex.RLock()
	defer loginMutex.RUnlock()
	return loginInProgress[userID]
}

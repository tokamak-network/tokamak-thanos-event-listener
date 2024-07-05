package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
)

type SlackData struct {
	Text string `json:"text"`
}

type SlackNotificationService struct {
	url        string
	numOfRetry int
	off        bool
}

func MakeSlackNotificationService(url string, numOfRetry int) *SlackNotificationService {
	return &SlackNotificationService{url: url, numOfRetry: numOfRetry, off: false}
}

func MakeDefaultSlackNotificationService() *SlackNotificationService {
	return &SlackNotificationService{url: os.Getenv("SLACK_URL"), numOfRetry: 5, off: os.Getenv("OFF") == "1"}
}

func (slackNotificationService *SlackNotificationService) Enable() {
	slackNotificationService.off = false
}

func (slackNotificationService *SlackNotificationService) Disable() {
	slackNotificationService.off = true
}

func (slackNotificationService *SlackNotificationService) Notify(title string, text string) error {
	if slackNotificationService.off {
		return nil
	}

	data := SlackData{
		Text: fmt.Sprintf("*%s*\n%s", title, text),
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", slackNotificationService.url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create a new client and execute our request
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and print the response
	body, _ := io.ReadAll(resp.Body)

	log.GetLogger().Infow("Response", "body", string(body))
	return nil
}

func (slackNotificationService *SlackNotificationService) NotifyWithReTry(title string, text string) {
	for i := 0; i < slackNotificationService.numOfRetry; i++ {
		err := slackNotificationService.Notify(title, text)
		if err == nil {
			break
		}
	}
}

package distribute

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type TextContext struct {
	Text string `json:"text"`
}

type Message struct {
	Timestamp string      `json:"timestamp"`
	Sign      string      `json:"sign"`
	MsgType   string      `json:"msg_type"`
	Content   TextContext `json:"content"`
}

type Notifier struct {
	Endpoint string
	Key      string
}

func GenSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func NewNotifier(endpoint, key string) (*Notifier, error) {
	return &Notifier{
		Endpoint: endpoint,
		Key:      key,
	}, nil
}

func (u *Notifier) SendMessage(message string) error {
	timestamp := time.Now().Unix()
	signature, err := GenSign(u.Key, timestamp)
	if err != nil {
		return err
	}
	msg := &Message{
		Timestamp: fmt.Sprintf("%d", timestamp),
		Sign:      signature,
		MsgType:   "text",
		Content: TextContext{
			Text: message,
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		u.Endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		log.Printf("new request error: %v\n", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

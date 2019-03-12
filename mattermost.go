package main

import (
    "encoding/json"
    "net/http"
    "strings"
)

type Webhook struct {
    URL         string
    IconURL     string
    Username    string
    Channel     string
}

type Message struct {
    Webhook
    Text        string
}

func (m *Message) Send() error {
    messageJSON, err := json.Marshal(m)
    if err != nil {
        return err
    }

    body := strings.NewReader(string(messageJSON))
    req, err := http.NewRequest("POST", m.URL, body)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }

    defer resp.Body.Close()
    return nil
}

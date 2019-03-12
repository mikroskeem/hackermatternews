package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
)

var (
    unauthorizedError = errors.New("Unauthorized")
)

type Client struct {
    httpClient *http.Client
    hnUrl string
}

func NewClient() Client {
    return Client{
        httpClient: &http.Client{},
        hnUrl: "https://hacker-news.firebaseio.com/v0",
    }
}

func (c *Client) GetLatestItemId() (int, error) {
    resp, err := c.httpClient.Get(fmt.Sprintf("%s/maxitem.json", c.hnUrl))
    if err != nil {
        return -1, err
    }

    defer resp.Body.Close()

    // Check for status code
    if resp.StatusCode != 200 {
        // TODO
        return -1, unauthorizedError
    }

    // Read body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return -1, err
    }

    // Parse
    var id int
    if err := json.Unmarshal(body, &id); err != nil {
        return -1, err
    }

    return id, nil
}

func (c *Client) GetItem(id int) (*HNItem, error) {
    resp, err := c.httpClient.Get(fmt.Sprintf("%s/item/%d.json", c.hnUrl, id))
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    // Check for status code
    if resp.StatusCode != 200 {
        // TODO
        return nil, unauthorizedError
    }

    // Read body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Parse
    var item HNItem
    if err := json.Unmarshal(body, &item); err != nil {
        return nil, err
    }

    return &item, nil
}

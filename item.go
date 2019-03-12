package main

type HNItem struct {
    Id      int     `json:"id"`
    Deleted bool    `json:"deleted"`
    Type    string  `json:"type"`
    By      string  `json:"by"`
    Time    uint64  `json:"time"`
    Text    string  `json:"text"`
    Url     string  `json:"url"`
    Title   string  `json:"title"`
}

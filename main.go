package main

import (
    "bytes"
    "fmt"
    "log"
    "net/url"
    "os"
    "os/signal"
    "text/template"
    "time"

    // Remote
    "github.com/davecgh/go-spew/spew"
    "github.com/lunny/html2md"
)

var (
    storyTemplate *template.Template
    askTemplate *template.Template
    webhookUrl string
)

func processAll(hnClient *Client, fromExcl int, toIncl int) {
    ticker := time.NewTicker(5 * time.Second)
    for id := fromExcl + 1; id <= toIncl; id++ {
        // fetch
        log.Printf("fetching item with id %d", id)
        item, err := hnClient.GetItem(id)
        if err != nil {
            log.Print("failed to get latest item: ", err)
            continue
        }

        // process
        spew.Dump(item)

        if item.Deleted {
            log.Printf("skipping deleted item %d", id)
            continue
        }

        switch item.Type {
        case "story":
            isAsk := len(item.Text) > 0

            var formattedArticle bytes.Buffer
            if isAsk {
                askTemplate.Execute(&formattedArticle, item)
            } else {
                storyTemplate.Execute(&formattedArticle, item)
            }

            message := Message{
                Webhook: Webhook{
                    URL: webhookUrl,
                },
                Text: formattedArticle.String(),
            }

            if err := message.Send(); err != nil {
                log.Print("failed to post news to mattermost: ", err)
            }
        case "comment":
            log.Print("someone commented on something, nobody cares")
        }

        // pause
        <- ticker.C
    }
    ticker.Stop()
}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    hnClient := NewClient()
    log.Print("hello world")

    // load configuration
    config, err := loadConfig("config.json")
    if err != nil {
        log.Panic("failed to load configuration", err)
    }

    if u, err := url.Parse(config.WebhookUrl); err != nil {
        log.Panic("invalid webhook url", err)
    } else {
        serializedUrl := u.String()
        if len(serializedUrl) < 1 {
            log.Panic("empty webhook url")
        }
        webhookUrl = serializedUrl
    }

    funcMap := template.FuncMap{
        "formatTime": func (tm uint64) string {
            return fmt.Sprintf("%s", time.Unix(int64(tm), 0))
        },
        "unHtmlify": func (html string) string {
            return html2md.Convert(html)
        },
    }

    storyTmpl, _ := template.New("story").Funcs(funcMap).Parse(
        "**{{.Title}}**\n" +
        "\n" +
        "**[Article]({{.Url}}), [Comments](https://news.ycombinator.com/item?id={{.Id}})**\n" +
        "**By:** [{{.By}}](https://news.ycombinator.com/user?id={{.By}})\n" +
        "**At:** _{{formatTime .Time}}_\n",
    )

    askTmpl, _ := template.New("ask").Funcs(funcMap).Parse(
        "**{{.Title}}**\n" +
        "{{unHtmlify .Text}}\n" +
        "\n" +
        "**Article, [Comments](https://news.ycombinator.com/item?id={{.Id}})**\n" +
        "**By:** [{{.By}}](https://news.ycombinator.com/user?id={{.By}})\n" +
        "**At:** _{{formatTime .Time}}_\n",
    )

    storyTemplate = storyTmpl
    askTemplate = askTmpl


    var latestId int = -1
    var currentId int = -1

    // start ticker to poll for updates
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for {
            log.Print("checking for new items...")
            if id, err := hnClient.GetLatestItemId(); err != nil {
                log.Print("failed to get latest item id: ", err)
                continue
            } else if latestId == id {
                log.Print("no updates this time")
                continue
            } else {
                if currentId == -1 {
                    log.Printf("setting current & latest id to %d", id)
                    currentId = id
                }

                latestId = id
            }

            // process items
            processAll(&hnClient, currentId, latestId)
            currentId = latestId

            // pause
            <- ticker.C
        }
    }()

    <- interrupt
    ticker.Stop()
}

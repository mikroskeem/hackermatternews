package main

import (
    "bytes"
    "fmt"
    "html"
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

func process(hnClient *Client, id int) {
    // fetch
    log.Printf("fetching item with id %d", id)
    item, err := hnClient.GetItem(id)
    if err != nil {
        log.Print("failed to get latest item: ", err)
        return
    }

    // process
    spew.Dump(item)

    if item.Deleted {
        log.Printf("skipping deleted item %d", id)
        return
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
}

func processAll(hnClient *Client, fromExcl int, toIncl int) {
    if fromExcl == toIncl {
        process(hnClient, toIncl)
        return
    }

    ticker := time.NewTicker(5 * time.Second)
    for id := fromExcl + 1; id <= toIncl; id++ {
        process(hnClient, id)

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

    // load state
    state, err := loadState("state.json")
    if err != nil {
        log.Panic("failed to load state!", err)
    }

    // set up templates
    funcMap := template.FuncMap{
        "formatTime": func (tm uint64) string {
            return fmt.Sprintf("%s", time.Unix(int64(tm), 0))
        },
        "unHtmlify": func (rawHtml string) string {
            return html.UnescapeString(html2md.Convert(rawHtml))
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

    // start ticker to poll for updates
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for {
            log.Print("checking for new items...")
            if id, err := hnClient.GetLatestItemId(); err != nil {
                log.Print("failed to get latest item id: ", err)
                <- ticker.C
                continue
            } else if state.LastLatestId == id {
                log.Print("no updates this time")
                <- ticker.C
                continue
            } else {
                if state.LastId == -1 {
                    log.Printf("setting current & latest id to %d", id)
                    state.LastId = id
                }

                state.LastLatestId = id
            }

            // process items
            processAll(&hnClient, state.LastId, state.LastLatestId)
            state.LastId = state.LastLatestId

            // pause
            <- ticker.C
        }
    }()

    // wait for interrupt signal
    <- interrupt
    ticker.Stop()

    // save state on exit
    log.Print("saving state")
    if err := state.save("state.json"); err != nil {
        log.Panic("failed to save state", err)
    }
    log.Print("done, exiting")
}

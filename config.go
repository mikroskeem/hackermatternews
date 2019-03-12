package main

import (
    "bufio"
    "encoding/json"
    "io/ioutil"
    "os"
)

type Configuration struct {
    WebhookUrl string `json:"mm_webhook_url"`
}

func loadConfig(fileName string) (*Configuration, error) {
    // check if configuration file exists
    if _, err := os.Stat(fileName); err != nil {
        // write sample one
        if os.IsNotExist(err) {
            sampleConfig := Configuration{}
            serialized, _ := json.MarshalIndent(sampleConfig, "", "    ")

            ioutil.WriteFile(fileName, serialized, 0600)

            // handled in program code
            return &sampleConfig, nil
        } else {
            // no idea what happened
            return nil, err
        }
    }

    if file, err := os.Open(fileName); err != nil {
        // failed to open
        return nil, err
    } else {
        defer file.Close()

        var deserializedConfig Configuration
        reader := bufio.NewReader(file)
        raw, err := ioutil.ReadAll(reader)
        if err != nil {
            return nil, err
        }

        if err := json.Unmarshal(raw, &deserializedConfig); err != nil {
            return nil, err
        }

        return &deserializedConfig, err
    }
}

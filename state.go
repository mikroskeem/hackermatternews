package main

import (
    "encoding/json"
    "io/ioutil"
    "os"
)

type State struct {
    LastId int `json:"last_id"`
    LastLatestId int `json:"last_latest_id"`
}

func NewState() State {
    return State{
        LastId: -1,
        LastLatestId: -1,
    }
}

func loadState(fileName string) (*State, error) {
    // check if configuration file exists
    if _, err := os.Stat(fileName); err != nil {
        // write sample one
        if os.IsNotExist(err) {
            newState := NewState()

            // handled in program code
            return &newState, nil
        } else {
            // no idea what happened
            return nil, err
        }
    }

    raw, err := ioutil.ReadFile(fileName)
    if err != nil {
        return nil, err
    }

    deserializedState := NewState()
    if err := json.Unmarshal(raw, &deserializedState); err != nil {
        return nil, err
    }

    return &deserializedState, nil
}

func (state *State) save(fileName string) (error) {
    serialized, _ := json.MarshalIndent(state, "", "    ")
    return ioutil.WriteFile(fileName, serialized, 0644)
}

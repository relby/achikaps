package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

type errorResp struct {
	Error string `json:"error"`
}

func main() {
	err := fmt.Errorf("1: %w", errors.New("2"))
	resp, err := json.Marshal(errorResp{Error: err.Error()})
	spew.Dump(resp)
	spew.Dump(err)
}
// Copyright 2017 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type globomapClient struct {
	Hostname string
}

type globomapPayload map[string]interface{}

type globomapProperty struct {
	name        string
	description string
	value       interface{}
}

type globomapResponse struct {
	Message string `json:"message"`
}

func (g *globomapClient) Post(ops []operation) error {
	path := "/v1/updates"
	body := g.body(ops)
	if body == nil {
		return errors.New("No events to post")
	}
	resp, err := g.doRequest(path, body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	if config.dry {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	var data globomapResponse
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println(data.Message)
	return nil
}

func (g *globomapClient) doRequest(path string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, g.Hostname+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	if config.dry {
		data, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s\n", data)
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Status:     "OK",
		}
		return resp, nil
	}
	return client.Do(req)
}

func (g *globomapClient) body(ops []operation) io.Reader {
	data := []globomapPayload{}
	for _, op := range ops {
		var payload *globomapPayload
		if op.docType == "collections" {
			payload = op.toDocument()
		} else {
			payload = op.toEdge()
		}
		if payload != nil {
			data = append(data, *payload)
		}
	}
	if len(data) == 0 {
		return nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bytes.NewReader(b)
}

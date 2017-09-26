// Copyright 2017 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type tsuruClient struct {
	Hostname string
	Token    string
}

type event struct {
	Target struct {
		Type  string
		Value string
	}
	StartTime time.Time
	EndTime   time.Time
}

type eventFilter struct {
	Kindname string
}

func (t *tsuruClient) EventList(f eventFilter) ([]event, error) {
	path := "/events"
	resp, err := t.doRequest(path + "?" + f.Apply().Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var events []event
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (t *tsuruClient) doRequest(path string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, t.Hostname+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "b "+t.Token)
	return client.Do(req)
}

func (f *eventFilter) Apply() url.Values {
	v := url.Values{}
	if f.Kindname != "" {
		v.Set("kindname", f.Kindname)
	}
	return v
}

// Copyright 2017 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"

	"gopkg.in/check.v1"
)

func newEvent(kind, value string) event {
	parts := strings.Split(kind, ".")
	e := event{}
	e.Target.Type = parts[0]
	e.Target.Value = value
	e.Kind.Name = kind
	return e
}

func (s *S) TestUpdateCmdRun(c *check.C) {
	tsuruServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/events":
			events := []event{
				newEvent("app.create", "myapp1"),
				newEvent("app.delete", "myapp2"),
				newEvent("pool.update", "pool1"),
			}
			json.NewEncoder(w).Encode(events)
		case "/apps/myapp1":
			json.NewEncoder(w).Encode(app{Name: "myapp1", Pool: "pool1"})
		case "/pools":
			json.NewEncoder(w).Encode([]pool{{Name: "pool1"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer tsuruServer.Close()
	os.Setenv("TSURU_HOSTNAME", tsuruServer.URL)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		c.Assert(r.Method, check.Equals, http.MethodPost)
		c.Assert(r.URL.Path, check.Equals, "/v1/updates")

		decoder := json.NewDecoder(r.Body)
		var data []globomapPayload
		err := decoder.Decode(&data)
		c.Assert(err, check.IsNil)
		defer r.Body.Close()
		c.Assert(data, check.HasLen, 5)

		sortPayload(data)
		el, ok := data[0]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[0]["action"], check.Equals, "UPDATE")
		c.Assert(data[0]["collection"], check.Equals, "tsuru_app")
		c.Assert(data[0]["type"], check.Equals, "collections")
		c.Assert(data[0]["key"], check.Equals, "tsuru_myapp1")
		c.Assert(el["name"], check.Equals, "myapp1")

		el, ok = data[1]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[1]["action"], check.Equals, "DELETE")
		c.Assert(data[1]["collection"], check.Equals, "tsuru_app")
		c.Assert(data[1]["type"], check.Equals, "collections")
		c.Assert(data[1]["key"], check.Equals, "tsuru_myapp2")
		c.Assert(el["name"], check.Equals, "myapp2")

		el, ok = data[2]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[2]["action"], check.Equals, "UPDATE")
		c.Assert(data[2]["collection"], check.Equals, "tsuru_pool")
		c.Assert(data[2]["type"], check.Equals, "collections")
		c.Assert(data[2]["key"], check.Equals, "tsuru_pool1")
		c.Assert(el["name"], check.Equals, "pool1")

		el, ok = data[3]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[3]["action"], check.Equals, "UPDATE")
		c.Assert(data[3]["collection"], check.Equals, "tsuru_pool_app")
		c.Assert(data[3]["type"], check.Equals, "edges")
		c.Assert(data[3]["key"], check.Equals, "tsuru_myapp1-pool")
		c.Assert(el["name"], check.Equals, "myapp1-pool")
		c.Assert(el["from"], check.Equals, "tsuru_app/tsuru_myapp1")
		c.Assert(el["to"], check.Equals, "tsuru_pool/tsuru_pool1")

		el, ok = data[4]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[4]["action"], check.Equals, "DELETE")
		c.Assert(data[4]["collection"], check.Equals, "tsuru_pool_app")
		c.Assert(data[4]["type"], check.Equals, "edges")
		c.Assert(data[4]["key"], check.Equals, "tsuru_myapp2-pool")
		c.Assert(el["name"], check.Equals, "myapp2-pool")
	}))
	defer server.Close()
	os.Setenv("GLOBOMAP_HOSTNAME", server.URL)
	setup(nil)

	cmd := &updateCmd{}
	cmd.Run()
	c.Assert(atomic.LoadInt32(&requests), check.Equals, int32(1))
}

func (s *S) TestUpdateCmdRunWithMultipleEventsPerKind(c *check.C) {
	tsuruServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/events":
			events := []event{
				newEvent("app.update", "myapp1"),
				newEvent("app.delete", "myapp1"),
				newEvent("app.create", "myapp2"),
				newEvent("app.update", "myapp2"),
				newEvent("pool.update", "pool1"),
				newEvent("pool.delete", "pool1"),
				newEvent("pool.create", "pool1"),
				newEvent("pool.create", "pool2"),
				newEvent("pool.delete", "pool2"),
			}
			json.NewEncoder(w).Encode(events)
		case "/apps/myapp1":
			json.NewEncoder(w).Encode(app{Name: "myapp1", Pool: "pool1"})
		case "/apps/myapp2":
			json.NewEncoder(w).Encode(app{Name: "myapp2", Pool: "pool1"})
		case "/pools":
			json.NewEncoder(w).Encode([]pool{{Name: "pool1"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer tsuruServer.Close()
	os.Setenv("TSURU_HOSTNAME", tsuruServer.URL)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		c.Assert(r.Method, check.Equals, http.MethodPost)
		c.Assert(r.URL.Path, check.Equals, "/v1/updates")

		decoder := json.NewDecoder(r.Body)
		var data []globomapPayload
		err := decoder.Decode(&data)
		c.Assert(err, check.IsNil)
		defer r.Body.Close()
		c.Assert(data, check.HasLen, 6)

		sortPayload(data)
		el, ok := data[0]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[0]["action"], check.Equals, "DELETE")
		c.Assert(data[0]["collection"], check.Equals, "tsuru_app")
		c.Assert(data[0]["type"], check.Equals, "collections")
		c.Assert(data[0]["key"], check.Equals, "tsuru_myapp1")
		c.Assert(el["name"], check.Equals, "myapp1")

		el, ok = data[1]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[1]["action"], check.Equals, "UPDATE")
		c.Assert(data[1]["collection"], check.Equals, "tsuru_app")
		c.Assert(data[1]["type"], check.Equals, "collections")
		c.Assert(data[1]["key"], check.Equals, "tsuru_myapp2")
		c.Assert(el["name"], check.Equals, "myapp2")

		el, ok = data[2]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[2]["action"], check.Equals, "UPDATE")
		c.Assert(data[2]["collection"], check.Equals, "tsuru_pool")
		c.Assert(data[2]["type"], check.Equals, "collections")
		c.Assert(data[2]["key"], check.Equals, "tsuru_pool1")
		c.Assert(el["name"], check.Equals, "pool1")

		el, ok = data[3]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[3]["action"], check.Equals, "DELETE")
		c.Assert(data[3]["collection"], check.Equals, "tsuru_pool")
		c.Assert(data[3]["type"], check.Equals, "collections")
		c.Assert(data[3]["key"], check.Equals, "tsuru_pool2")
		c.Assert(el["name"], check.Equals, "pool2")

		el, ok = data[4]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[4]["action"], check.Equals, "DELETE")
		c.Assert(data[4]["collection"], check.Equals, "tsuru_pool_app")
		c.Assert(data[4]["type"], check.Equals, "edges")
		c.Assert(data[4]["key"], check.Equals, "tsuru_myapp1-pool")
		c.Assert(el["name"], check.Equals, "myapp1-pool")

		el, ok = data[5]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[5]["action"], check.Equals, "UPDATE")
		c.Assert(data[5]["collection"], check.Equals, "tsuru_pool_app")
		c.Assert(data[5]["type"], check.Equals, "edges")
		c.Assert(data[5]["key"], check.Equals, "tsuru_myapp2-pool")
		c.Assert(el["name"], check.Equals, "myapp2-pool")
		c.Assert(el["from"], check.Equals, "tsuru_app/tsuru_myapp2")
		c.Assert(el["to"], check.Equals, "tsuru_pool/tsuru_pool1")
	}))
	defer server.Close()
	os.Setenv("GLOBOMAP_HOSTNAME", server.URL)
	setup(nil)

	cmd := &updateCmd{}
	cmd.Run()
	c.Assert(atomic.LoadInt32(&requests), check.Equals, int32(1))
}

func (s *S) TestUpdateCmdRunNoRequestWhenNoEventsToPost(c *check.C) {
	tsuruServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/events" {
			json.NewEncoder(w).Encode([]event{})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer tsuruServer.Close()
	os.Setenv("TSURU_HOSTNAME", tsuruServer.URL)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.ExpectFailure("No request should have been done")
	}))
	defer server.Close()
	os.Setenv("GLOBOMAP_HOSTNAME", server.URL)
	setup(nil)

	cmd := &updateCmd{}
	cmd.Run()
	c.Assert(atomic.LoadInt32(&requests), check.Equals, int32(0))
}

func (s *S) TestUpdateCmdRunAppProperties(c *check.C) {
	tsuruServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/events":
			events := []event{
				newEvent("app.create", "myapp1"),
			}
			json.NewEncoder(w).Encode(events)
		case "/apps/myapp1":
			a := app{
				Name:        "myapp1",
				Description: "about my app",
				Tags:        []string{"tag1", "tag2"},
				Platform:    "go",
				Ip:          "myapp1.example.com",
				Cname:       []string{"myapp1.alias.com"},
				Router:      "galeb",
				Owner:       "me@example.com",
				TeamOwner:   "my-team",
				Teams:       []string{"team1", "team2"},
				Plan:        appPlan{Name: "large", Router: "galeb1", Memory: 1073741824, Swap: 0, Cpushare: 1024},
			}
			json.NewEncoder(w).Encode(a)
		}
	}))
	defer tsuruServer.Close()
	os.Setenv("TSURU_HOSTNAME", tsuruServer.URL)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		c.Assert(r.Method, check.Equals, http.MethodPost)
		c.Assert(r.URL.Path, check.Equals, "/v1/updates")

		decoder := json.NewDecoder(r.Body)
		var data []globomapPayload
		err := decoder.Decode(&data)
		c.Assert(err, check.IsNil)
		defer r.Body.Close()
		c.Assert(data, check.HasLen, 2)

		sortPayload(data)
		el, ok := data[0]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[0]["action"], check.Equals, "UPDATE")
		c.Assert(data[0]["collection"], check.Equals, "tsuru_app")
		c.Assert(data[0]["type"], check.Equals, "collections")
		c.Assert(data[0]["key"], check.Equals, "tsuru_myapp1")
		c.Assert(el["name"], check.Equals, "myapp1")
		props, ok := el["properties"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(props["description"], check.Equals, "about my app")
		c.Assert(props["tags"], check.Equals, "tag1, tag2")
		c.Assert(props["platform"], check.Equals, "go")
		c.Assert(props["addresses"], check.Equals, "myapp1.alias.com, myapp1.example.com")
		c.Assert(props["router"], check.Equals, "galeb")
		c.Assert(props["owner"], check.Equals, "me@example.com")
		c.Assert(props["team_owner"], check.Equals, "my-team")
		c.Assert(props["teams"], check.Equals, "team1, team2")
		c.Assert(props["plan_name"], check.Equals, "large")
		c.Assert(props["plan_router"], check.Equals, "galeb1")
		c.Assert(props["plan_memory"], check.Equals, "1073741824")
		c.Assert(props["plan_swap"], check.Equals, "0")
		c.Assert(props["plan_cpushare"], check.Equals, "1024")
		_, ok = el["properties_metadata"]
		c.Assert(ok, check.Equals, true)

		el, ok = data[1]["element"].(map[string]interface{})
		c.Assert(ok, check.Equals, true)
		c.Assert(data[1]["action"], check.Equals, "UPDATE")
		c.Assert(data[1]["collection"], check.Equals, "tsuru_pool_app")
		c.Assert(data[1]["type"], check.Equals, "edges")
		c.Assert(data[1]["key"], check.Equals, "tsuru_myapp1-pool")
		c.Assert(el["name"], check.Equals, "myapp1-pool")
		_, ok = el["properties"]
		c.Assert(ok, check.Equals, false)
		_, ok = el["properties_metadata"]
		c.Assert(ok, check.Equals, false)
	}))
	defer server.Close()
	os.Setenv("GLOBOMAP_HOSTNAME", server.URL)
	setup(nil)

	cmd := &updateCmd{}
	cmd.Run()
	c.Assert(atomic.LoadInt32(&requests), check.Equals, int32(1))
}

func (s *S) TestUpdateCmdRunIgnoresFailedEvents(c *check.C) {
	tsuruServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/events":
			failedEvent := newEvent("app.delete", "myapp1")
			failedEvent.Error = "something wrong happened"
			events := []event{
				newEvent("app.create", "myapp1"),
				failedEvent,
			}
			json.NewEncoder(w).Encode(events)
		case "/apps/myapp1":
			json.NewEncoder(w).Encode(app{})
		}
	}))
	defer tsuruServer.Close()
	os.Setenv("TSURU_HOSTNAME", tsuruServer.URL)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		c.Assert(r.Method, check.Equals, http.MethodPost)
		c.Assert(r.URL.Path, check.Equals, "/v1/updates")

		decoder := json.NewDecoder(r.Body)
		var data []globomapPayload
		err := decoder.Decode(&data)
		c.Assert(err, check.IsNil)
		defer r.Body.Close()
		c.Assert(data, check.HasLen, 2)
	}))
	defer server.Close()
	os.Setenv("GLOBOMAP_HOSTNAME", server.URL)
	setup(nil)

	cmd := &updateCmd{}
	cmd.Run()
	c.Assert(atomic.LoadInt32(&requests), check.Equals, int32(1))
}

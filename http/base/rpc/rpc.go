/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package rpc provides functions to for rpc server call
package rpc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	// fast json marshal/unmarshal
	jsoniter "github.com/json-iterator/go"
	"github.com/qbyyf/ontology/common/log"
	"github.com/qbyyf/ontology/http/base/common"
	berr "github.com/qbyyf/ontology/http/base/error"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type JReq struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      interface{}   `json:"id"`
}

func init() {
	mainMux.m = make(map[string]func([]interface{}) map[string]interface{})
}

//an instance of the multiplexer
var mainMux ServeMux

//multiplexer that keeps track of every function to be called on specific rpc call
type ServeMux struct {
	sync.RWMutex
	m               map[string]func([]interface{}) map[string]interface{}
	defaultFunction func(http.ResponseWriter, *http.Request)
}

//a function to register functions to be called for specific rpc calls
func HandleFunc(pattern string, handler func([]interface{}) map[string]interface{}) {
	mainMux.Lock()
	defer mainMux.Unlock()
	mainMux.m[pattern] = handler
}

//a function to be called if the request is not a HTTP JSON RPC call
func SetDefaultFunc(def func(http.ResponseWriter, *http.Request)) {
	mainMux.defaultFunction = def
}

// this is the function that should be called in order to answer an rpc call
// should be registered like "http.HandleFunc("/", httpjsonrpc.Handle)"
func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}
	//JSON RPC commands should be POSTs
	if r.Method != "POST" {
		if mainMux.defaultFunction != nil {
			log.Info("HTTP JSON RPC Handle - Method!=\"POST\"")
			mainMux.defaultFunction(w, r)
		} else {
			log.Warn("HTTP JSON RPC Handle - Method!=\"POST\"")
		}
		return
	}
	//check if there is Request Body to read
	if r.Body == nil {
		mainMux.RLock()
		if mainMux.defaultFunction != nil {
			log.Info("HTTP JSON RPC Handle - Request body is nil")
			mainMux.defaultFunction(w, r)
		} else {
			log.Warn("HTTP JSON RPC Handle - Request body is nil")
		}
		mainMux.RUnlock()
		return
	}
	var request JReq
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, common.MAX_REQUEST_BODY_SIZE))
	err := decoder.Decode(&request)
	if err != nil {
		log.Error("HTTP JSON RPC Handle - json.Unmarshal: ", err)
		return
	}
	if request.Method == "" {
		log.Error("HTTP JSON RPC Handle - method is not string: ")
		return
	}
	//get the corresponding function
	mainMux.RLock()
	function, ok := mainMux.m[request.Method]
	mainMux.RUnlock()
	if ok {
		response := function(request.Params)
		data, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"error":   response["error"],
			"desc":    response["desc"],
			"result":  response["result"],
			"id":      request.ID,
		})
		if err != nil {
			log.Error("HTTP JSON RPC Handle - json.Marshal: ", err)
			return
		}
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	} else {
		//if the function does not exist
		log.Warn("HTTP JSON RPC Handle - No function to call for ", request.Method)
		data, err := json.Marshal(map[string]interface{}{
			"error": berr.INVALID_METHOD,
			"result": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
				"data":    "The called method was not found on the server",
			},
			"id": request.ID,
		})
		if err != nil {
			log.Error("HTTP JSON RPC Handle - json.Marshal: ", err)
			return
		}
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	}
}

// Call sends RPC request to server
func Call(address string, method string, id interface{}, params []interface{}) ([]byte, error) {
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     id,
		"params": params,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Marshal JSON request: %v\n", err)
		return nil, err
	}

	resp, err := http.Post(address, "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GET response: %v\n", err)
		return nil, err
	}

	return body, nil
}

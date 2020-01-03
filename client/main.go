// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"

	// "io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	// "github.com/septianw/jas/common"

	"github.com/gorilla/websocket"
)

type Item struct {
	Qty  int    `json:"qty"`
	Name string `json:"name"`
}

type Order struct {
	Id    int    `json:"id"`
	Items []Item `json:"item"`
}

type Orders []Order

var addr = flag.String("addr", "127.0.0.1:4519", "http service address")
var id = 0

func Getdata() Orders {
	var orders Orders
	var items []Item
	qtypool := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	namepool := []string{"ayam geprek", "chicken katsu", "cordon bleu", "paket 1", "paket 2", "es teh manis", "teh manis", "teh tawar", "frestea", "lemonade", "lemon squash"}
	rand.Seed(2048)

	for oi := 0; oi < rand.Intn(45); oi++ {
		for i := 0; i < rand.Intn(15); i++ {
			item := Item{rand.Intn(len(qtypool)), namepool[rand.Intn(len(namepool))]}
			log.Printf("%+v\n", item)

			items = append(items, item)
		}
		orders = append(orders, Order{id + 1, items})
	}

	return orders
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	d, _ := json.Marshal(Getdata())
	log.Println(string(d))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/api/v1/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// var file string = "msg.json"
			// var osf *os.File
			// _, err := os.Stat(file)
			// if os.IsNotExist(err) {
			// 	osf, err = os.Create(file)
			// 	common.ErrHandler(err)
			// }
			// defer osf.Close()
			// data, err := ioutil.ReadFile(file)
			// common.ErrHandler(err)
			data, _ := json.Marshal(Getdata())

			err = c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

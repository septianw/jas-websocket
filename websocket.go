package main

import (
	"fmt"
	"io/ioutil"
	"log"

	// "net"
	"net/http"
	"os"

	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/septianw/jas/common"
	// "github.com/spf13/viper"
)

var datafile = "/tmp/hotfile.txt"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

func Bootstrap() {
	fmt.Println("Module websocket bootstrap.")
}

func Router(r *gin.Engine) {
	r.GET("/api/v1/ws", func(c *gin.Context) {
		WsHandler(c.Writer, c.Request)
	})

	r.StaticFS("/ui", FS(false))
}

func checkWhitelist(r *http.Request) bool {
	// var out bool = false
	// // var out bool = true

	// rt := common.ReadRuntime()
	// viper.AddConfigPath(rt.ConfigLocation)
	// viper.ReadInConfig()

	// whitelist := viper.GetStringSlice("websocket.whitelist")
	// log.Printf("\n%+v\n", whitelist)
	// h, _, err := net.SplitHostPort(r.RemoteAddr)
	// common.ErrHandler(err)
	// ip := net.ParseIP(h)
	// if ip.To4() == nil {
	// 	s := strings.Split(r.RemoteAddr, ":")
	// 	h = strings.Join(s[:len(s)-1], ":")
	// } else {
	// 	h = strings.Split(r.RemoteAddr, ":")[0]
	// }

	// for _, element := range whitelist {
	// 	log.Printf("Plain Remote: %+v\n", r.RemoteAddr)
	// 	log.Printf("Remote: %+v\n", h)
	// 	log.Printf("Element: %+v\n", element)
	// 	if strings.Compare(h, element) == 0 {
	// 		out = true
	// 		break
	// 	}
	// }

	// return out
	return true
}

func Reader(c *websocket.Conn, readHandler func(int, []byte) error) {
	for {
		mType, p, err := c.ReadMessage()
		common.ErrHandler(err)

		log.Println(string(p))

		err = readHandler(mType, p)
		common.ErrHandler(err)
	}
}

func NotifyFromFS(
	c *websocket.Conn,
	mType int,
	p []byte,
	notifyAction func(*websocket.Conn, int, []byte) error,
) {
	watcher, err := fsnotify.NewWatcher()
	common.ErrHandler(err)
	defer watcher.Close()
	log.Println("watcher run.")
	log.Println(mType, p, c)

	done := make(chan bool)
	go func() {
		for {
			log.Println("goroutine run.")
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					fc, err := ioutil.ReadFile(event.Name)
					log.Println(string(fc))
					common.ErrHandler(err)
					log.Println(mType, p, c)
					notifyAction(c, 1, fc)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(datafile)
	common.ErrHandler(err)
	<-done
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("\n%+v\n", strings.Split(r.RemoteAddr, ":")[0])
	wsUpgrader.CheckOrigin = checkWhitelist
	var mType int
	var p []byte

	c, err := wsUpgrader.Upgrade(w, r, nil)
	common.ErrHandler(err)

	// mType, msg, err := c.ReadMessage()
	// log.Println(mType, string(msg), err)

	if _, err := os.Stat(datafile); os.IsNotExist(err) {
		ioutil.WriteFile(datafile, []byte(" "), 0664)
	}
	hfc, err := ioutil.ReadFile(datafile)
	common.ErrHandler(err)

	err = c.WriteMessage(1, hfc)
	common.ErrHandler(err)
	NotifyFromFS(c, mType, p, func(c *websocket.Conn, mType int, p []byte) error {
		log.Println("notify action run.")
		err := c.WriteMessage(mType, p)
		log.Println(mType, p)
		common.ErrHandler(err)
		return nil
	})

	log.Println(mType, p, c)

	Reader(c, func(mType int, p []byte) error {
		// websocket.TextMessage
		log.Println(string(p), mType)
		// return nil
		return c.WriteMessage(mType, p)
	})

}

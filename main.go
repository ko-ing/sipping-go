package main

import (
	"github.com/gin-gonic/gin"
	m "gopkg.in/olahol/melody.v1"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type GopherInfo struct {
	ID, X, Y string
}

func main() {
	router := gin.Default()
	melody := m.New()
	lock := new(sync.Mutex)
	counter := 0
	gophers := make(map[*m.Session]*GopherInfo)

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "main.html")
	})

	router.GET("/ws", func(c *gin.Context) {
		melody.HandleRequest(c.Writer, c.Request)
	})

	melody.HandleConnect(func(s *m.Session) {
		lock.Lock()
		for _, info := range gophers {
			s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
		}
		gophers[s] = &GopherInfo{strconv.Itoa(counter), "0", "0"}
		s.Write([]byte("iam " + gophers[s].ID))
		counter += 1
		lock.Unlock()
	})

	melody.HandleDisconnect(func(s *m.Session) {
		lock.Lock()
		melody.BroadcastOthers([]byte("dis "+gophers[s].ID), s)
		delete(gophers, s)
		lock.Unlock()
	})

	melody.HandleMessage(func(s *m.Session, msg []byte) {
		p := strings.Split(string(msg), " ")
		lock.Lock()
		info := gophers[s]
		if len(p) == 2 {
			info.X = p[0]
			info.Y = p[1]
			melody.BroadcastOthers([]byte("set "+info.ID+" "+info.X+" "+info.Y), s)
		}
		lock.Unlock()
	})

	router.Run(":5000")
}

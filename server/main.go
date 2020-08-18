package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tucher/autosettings"

	"net/http"
	"sync"

	"mesh-simulator/meshpeer"
	"mesh-simulator/meshsim"
)

type config struct {
	DEBUG          bool
	LogFile        string `autosettings:"logfile full path or stdout"`
	HTTPAddress    string `autosettings:"address and port for http mode"`
	HistorySeconds int
}

func (*config) Default() autosettings.Defaultable {
	return &config{
		DEBUG:       true,
		LogFile:     "stdout",
		HTTPAddress: "0.0.0.0:8088",
	}
}

func getLogger(name string) *log.Logger {
	var logWriter io.Writer
	if name == "stdout" {
		logWriter = os.Stdout
	} else {
		var err error
		logWriter, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println(err)
			logWriter = os.Stdout
		}
	}
	return log.New(logWriter, "", log.Llongfile|log.Ldate|log.Ltime)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {

	conf := &config{}
	autosettings.ReadConfig(conf)
	logger := getLogger(conf.LogFile)
	logger.Println("Started")
	r := gin.Default()
	r.Use(cors.Default())

	crowdSimulator := meshsim.New(logger)
	for i := 0; i < 10; i++ {
		npc := meshpeer.NewSimplePeer1(log.New(os.Stdout, "SIMPLE PEER", log.LstdFlags))
		crowdSimulator.AddActor(npc, [2]float64{53.904153, 27.556925})
	}
	r.GET("/state_overview", func(c *gin.Context) {
		logger.Println("here")
		c.JSON(http.StatusOK, crowdSimulator.GetOverview())
	})

	wsMutex := sync.RWMutex{}
	allConns := make(map[string]*wsClient)

	r.GET("/ws_rpc", func(c *gin.Context) {
		latlon := [2]float64{
			53.904153,
			27.556925,
		}
		if lat, err := strconv.ParseFloat(c.Query("lat"), 32); err == nil {
			latlon[0] = lat
		}
		if lon, err := strconv.ParseFloat(c.Query("lon"), 32); err == nil {
			latlon[1] = lon
		}
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Println("Failed to set websocket upgrade: ", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		newConn := &wsClient{
			conn:       conn,
			outChannel: make(chan []byte),
			inChannel:  make(chan []byte),
		}
		newConn.meshPeer = meshpeer.NewRPCPeer(newConn.inChannel, newConn.outChannel, log.New(os.Stdout, "RPC PEER", log.LstdFlags))
		newConn.meshPeerID = crowdSimulator.AddActor(newConn.meshPeer, latlon)
		wsMutex.Lock()
		allConns[conn.RemoteAddr().String()] = newConn
		logger.Println("WS connections count: ", len(allConns))
		wsMutex.Unlock()

		newConn.run(logger)

		close(newConn.inChannel)

		crowdSimulator.RemoveActor(newConn.meshPeerID)
		wsMutex.Lock()
		delete(allConns, conn.RemoteAddr().String())
		logger.Println("WS connections count: ", len(allConns))
		wsMutex.Unlock()

	})
	r.StaticFile("/", "./static/viewer.html")
	r.Static("/static", "./static")

	r.Run(conf.HTTPAddress)
}

type wsClient struct {
	conn       *websocket.Conn
	outChannel chan []byte
	inChannel  chan []byte

	meshPeerID meshsim.NetworkID
	meshPeer   *meshpeer.RPCPeer
}

func (cl *wsClient) run(logger *log.Logger) {
	done := make(chan bool)
	defer func() {
		cl.conn.Close()
		close(done)
		close(cl.inChannel)
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case n := <-cl.outChannel:
				cl.conn.WriteMessage(websocket.TextMessage, n)
			}
		}

	}()

	for {
		_, msg, err := cl.conn.ReadMessage()
		if err != nil {
			return
		}
		cl.inChannel <- msg
	}
}

func appendToFile(filename string, data []byte) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(data); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

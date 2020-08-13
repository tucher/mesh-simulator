package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tucher/autosettings"

	"encoding/json"
	"net/http"
	"sync"

	crowd "mesh-simulator/crowd_model"
)

type config struct {
	DEBUG          bool
	LogFile        string `autosettings:"logfile full path or stdout"`
	HTTPAddress    string `autosettings:"address and port for http mode"`
	HistorySeconds int
}

func (*config) Default() autosettings.Defaultable {
	return &config{
		DEBUG:          true,
		LogFile:        "stdout",
		HTTPAddress:    "0.0.0.0:8088",
		HistorySeconds: 60,
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

	crowdSimulator := crowd.New(logger)
	crowdSimulator.AddNPC(10, [2]float64{53.904153, 27.556925})
	r.GET("/state_overview", func(c *gin.Context) {
		// ttt := 1 / 0
		// logger.Println(ttt)
		logger.Println("here")
		c.JSON(http.StatusOK, crowdSimulator.GetOverview())
	})

	wsMutex := sync.RWMutex{}
	allConns := make(map[string]*WSClient)

	r.GET("/ws_rpc", func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Println("Failed to set websocket upgrade: ", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		newConn := WSClient{conn, make(chan jsonCommand)}
		wsMutex.Lock()
		allConns[conn.RemoteAddr().String()] = &newConn
		logger.Println("WS connections count: ", len(allConns))
		wsMutex.Unlock()

		newConn.run(logger)
		wsMutex.Lock()
		delete(allConns, conn.RemoteAddr().String())
		logger.Println("WS connections count: ", len(allConns))
		wsMutex.Unlock()

	})
	r.StaticFile("/", "./static/viewer.html")
	r.Static("/static", "./static")

	r.Run(conf.HTTPAddress)
}

type WSClient struct {
	conn       *websocket.Conn
	outChannel chan jsonCommand
}

type jsonCommand struct {
	Cmd  string
	Data json.RawMessage
}

func (th *jsonCommand) SerData(data interface{}) {
	th.Data, _ = json.Marshal(data)
}

func (th *jsonCommand) GetData(obj interface{}) error {
	return json.Unmarshal(th.Data, obj)
}

func (cl *WSClient) run(logger *log.Logger) {
	done := make(chan bool)
	defer func() {
		cl.conn.Close()
		close(done)
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case n := <-cl.outChannel:
				cl.conn.WriteJSON(n)
			}
		}

	}()

	// go func() {
	// 	ticker := time.NewTicker(time.Second)
	// 	for {
	// 		select {
	// 		case <-done:
	// 			return
	// 		case <-ticker.C:
	// 			t := jsonCommand{Cmd: "something_for_client"}
	// 			t.SerData(map[string]interface{}{"ts": time.Now().Unix()})
	// 			cl.outChannel <- t
	// 		}
	// 	}
	// }()

	for {
		_, msg, err := cl.conn.ReadMessage()
		if err != nil {
			return
		}
		cmd := jsonCommand{}
		if json.Unmarshal(msg, &cmd) != nil {
			return
		}
		switch cmd.Cmd {
		default:
			logger.Printf("WS MESSAGE: %+v", cmd)
		}
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

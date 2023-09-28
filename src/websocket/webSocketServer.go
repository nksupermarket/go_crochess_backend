package domain_websocket

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"nhooyr.io/websocket"
)

const (
	PingPeriod = time.Second * 30
)

var clientID = 0

type WebSocketServer struct {
	Conns      map[int]Client
	register   chan Client
	unregister chan Client
	router     WebSocketRouter
}

func NewWebSocketServer(r WebSocketRouter) *WebSocketServer {
	return &WebSocketServer{
		Conns:      make(map[int]Client),
		register:   make(chan Client),
		unregister: make(chan Client),
		router:     r,
	}
}

func (s *WebSocketServer) Run() {
	select {
	case client := <-s.unregister:
		s.unregisterClient(client)
	}
}

func (s *WebSocketServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	wsConfig := websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{"*"},
		CompressionMode:    websocket.CompressionDisabled,
	}
	conn, err := websocket.Accept(w, r, &wsConfig)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	var client Client
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		client = NewClient(clientID, conn, s)
		clientID += 1
	} else {
		id, err := strconv.Atoi(uid)
		if err != nil {
			log.Printf("error parsing user id when connecting: %v", err)
			conn.Close(websocket.StatusInternalError, fmt.Sprintf(`"%s" is not a valid uid`, uid))
			return
		}
		client = NewClient(id, conn, s)
	}

	go client.ReadPump(r.Context())
	go client.WritePump(r.Context())

	s.registerClient(client)
}

func (s *WebSocketServer) registerClient(client Client) {
	s.Conns[client.GetID()] = client
}

func (s *WebSocketServer) unregisterClient(client Client) {
	if _, ok := s.Conns[client.GetID()]; ok {
		delete(s.Conns, client.GetID())
	}
}

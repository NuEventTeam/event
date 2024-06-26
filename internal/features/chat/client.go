package chat

import (
	"context"
	"github.com/NuEventTeam/events/internal/features/chat/chat_features"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var DB *database.Database

const (
	// Time allowed to write a message to the peer.
	writeWait = 30 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ClientList map[int64]*Client

type Client struct {
	ClientId    int64
	EventId     int64
	Conn        *websocket.Conn
	Manager     *Manager
	SendMsgChan chan Message
}

func NewClient(userId, eventId int64, m *Manager, conn *websocket.Conn) *Client {
	return &Client{
		ClientId: userId,
		EventId:  eventId,
		Conn:     conn,
		Manager:  m,

		SendMsgChan: make(chan Message, 100),
	}
}

func (c *Client) writeMessage(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("stopping chat")
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.SendMsgChan:
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message.Payload)
			if err != nil {
				log.Printf("[%d] err :%v\n", c.ClientId, err)
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *Client) readMessage(ctx context.Context) {

	defer func() {
		c.Manager.Unregister(c)
		c.Conn.Close()
		log.Println("stopping chat")

	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			log.Println(err)
			break
		}

		msg, err := chat_features.SaveMessage(DB.GetDb(), c.EventId, c.ClientId, string(payload))
		if err != nil {
			log.Println(err)
			return
		}

		js, err := sonic.ConfigFastest.Marshal(msg)

		c.Manager.messageChan <- Message{
			EventId: ctx.Value("eventId").(int64),
			Payload: js,
			From:    c.ClientId,
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Println(reason, status)
	},
}

func ServeWs(manager *Manager, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	userId := r.Context().Value("userId").(int64)
	eventId := r.Context().Value("eventId").(int64)
	client := NewClient(userId, eventId, manager, conn)

	client.Manager.Register(client)

	go client.writeMessage(r.Context())
	go client.readMessage(r.Context())
}

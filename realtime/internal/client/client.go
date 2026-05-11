package client

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/config"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
)

const (
	sendBufSize = 256
)

// Client — одно WebSocket-соединение.
type Client struct {
	ID       string
	Name     string
	Role     string
	RoomCode string
	UserID   string

	conn   *websocket.Conn
	Send   chan message.OutgoingMessage
	closed bool
	mu     sync.Mutex

	cfg    *config.Config
	logger *slog.Logger

	OnMessage    func(c *Client, msg message.IncomingMessage)
	OnDisconnect func(c *Client)
}

// New создаёт нового клиента.
func New(conn *websocket.Conn, cfg *config.Config, logger *slog.Logger) *Client {
	return &Client{
		ID:     uuid.NewString(),
		conn:   conn,
		Send:   make(chan message.OutgoingMessage, sendBufSize),
		cfg:    cfg,
		logger: logger.With("component", "client"),
	}
}

// Start запускает горутины read/write pump.
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

// readPump читает входящие сообщения из WebSocket и передаёт их в OnMessage.
func (c *Client) readPump() {
	defer func() {
		c.close()
		if c.OnDisconnect != nil {
			c.OnDisconnect(c)
		}
	}()

	c.conn.SetReadLimit(c.cfg.WSMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.cfg.WSPongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.cfg.WSPongWait))
		return nil
	})

	for {
		_, rawMsg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived,
			) {
				c.logger.Warn("неожиданное закрытие WS", "client_id", c.ID, "error", err)
			}
			break
		}

		var incoming message.IncomingMessage
		if err := json.Unmarshal(rawMsg, &incoming); err != nil {
			c.logger.Warn("невалидный JSON от клиента", "client_id", c.ID, "raw", string(rawMsg))
			c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, "невалидный формат сообщения"))
			continue
		}

		if incoming.Type == "" {
			c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, "поле type обязательно"))
			continue
		}

		if c.OnMessage != nil {
			c.OnMessage(c, incoming)
		}
	}
}

// writePump читает сообщения из канала Send и отправляет их клиенту.
func (c *Client) writePump() {
	ticker := time.NewTicker(c.cfg.WSPingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WSWriteWait))
			if !ok {
				// Канал закрыт — закрываем соединение.
				_ = c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				c.logger.Warn("ошибка отправки сообщения", "client_id", c.ID, "error", err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WSWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMsg помещает сообщение в очередь отправки.
func (c *Client) SendMsg(msg message.OutgoingMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.Send <- msg:
	default:
		// Буфер переполнен — клиент не успевает читать.
		c.logger.Warn("буфер клиента переполнен, отключаем", "client_id", c.ID)
		c.closeLocked()
	}
}

// close безопасно закрывает клиента.
func (c *Client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
}

func (c *Client) closeLocked() {
	if !c.closed {
		c.closed = true
		close(c.Send)
	}
}

// IsClosed возвращает true, если клиент отключён.
func (c *Client) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

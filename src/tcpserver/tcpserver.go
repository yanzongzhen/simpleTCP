package tcpserver

import (
	"context"
	logger "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

type TcpConn interface {
	Write([]byte) (int, error)
	Close() error
	SetDeviceId(deviceId string)
	GetDeviceId() string
	SetDeviceStatus(status string)
	GetDeviceStatus() int
	GetProperty() Property
}

type TimeoutFunc func(c TcpConn)
type MessageHandle func(c TcpConn, msg []byte)
type DeviceIDSetHook func(c TcpConn)
type Option func(c *conn)

const Open = 1
const Close = 0

var OpenState = []byte{1, 10}
var CloseState = []byte{1, 11}
var CloseResp = []byte{1, 12}
var OpenResp = []byte{1, 13}

type Property struct {
	DeviceID        string    `json:"deviceId"`
	Status          int       `json:"status"`
	LastMessageTime time.Time `json:"lastMessageTime"`
}

type conn struct {
	deviceId         string
	c                net.Conn
	mux              *sync.Mutex
	lastMessage      time.Time
	timeoutFunc      TimeoutFunc
	onMessage        MessageHandle
	onChangeDeviceId DeviceIDSetHook
	checkDuration    time.Duration
	ctx              context.Context
	cancel           context.CancelFunc
	initMsg          string
	receiveSeq       int
	status           int
}

func (c *conn) GetProperty() Property {
	var ret Property
	ret.DeviceID = c.deviceId
	ret.Status = c.status
	ret.LastMessageTime = c.lastMessage
	return ret
}

func (c *conn) Write(data []byte) (int, error) {
	c.mux.Lock()
	defer func() {
		c.mux.Unlock()

	}()
	return c.c.Write(data)
}

func (c *conn) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cancel()
	return c.c.Close()
}

// SetDeviceId
func (c *conn) SetDeviceId(deviceId string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.deviceId = deviceId
}

// GetDeviceId
func (c *conn) GetDeviceId() string {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.deviceId
}

// SetDeviceId
func (c *conn) SetDeviceStatus(status string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if status == "010D" || status == "010A" {
		c.status = Open
	} else if status == "010C" || status == "010B" {
		c.status = Close
	} else {
		logger.Error("not support cmd")
	}
}

// GetDeviceId
func (c *conn) GetDeviceStatus() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.status
}

func (c *conn) keepAlive() {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-c.ctx.Done():
			logger.Debug("keepAlive exit")
			t.Stop()
			return
		case <-t.C:
			if time.Now().Sub(c.lastMessage) > c.checkDuration {
				if c.timeoutFunc != nil {
					c.timeoutFunc(c)
				}
			}
		}
	}
}

// readLoop .
func (c *conn) readLoop() {
	for {
		select {
		case <-c.ctx.Done():
			logger.Debug("readLoop exit")
			return
		default:
			var line string
			buffer := make([]byte, 1024)
			recvLen, err := c.c.Read(buffer)
			if err != nil {
				continue
			}
			rawByte := buffer[:recvLen]
			logger.Debug(string(rawByte));
			line = strings.ReplaceAll(string(rawByte), "\r\n", "")
			// logger.Warn(rawByte, len(rawByte))
			c.mux.Lock()
			switch (line) {
			case "010A", "010D":
				c.status = Open;
				break
			case "010C", "010B":
				c.status = Close
				break
			}
			c.lastMessage = time.Now()
			c.mux.Unlock()
			// logger.Debugf("==> origin data [%s] [%d]", line, len(line))
			if c.receiveSeq == 0 {
				c.SetDeviceId(line)
				if c.onChangeDeviceId != nil {
					c.onChangeDeviceId(c)
				}
			}
			c.receiveSeq++
			c.onMessage(c, []byte(line))
		}
	}
}

func byteEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		current := a[i]
		if current != b[i] {
			return false
		}
	}
	return true
}

// initSendMsg .
func (c *conn) initSendMsg() {
	_, _ = c.Write([]byte(c.initMsg))
}

func NewTcpConn(c net.Conn, opts ...Option) TcpConn {
	ctx, cancel := context.WithCancel(context.Background())
	instance := &conn{
		c:             c,
		lastMessage:   time.Now(),
		mux:           new(sync.Mutex),
		ctx:           ctx,
		cancel:        cancel,
		checkDuration: time.Second * 30,
	}
	for _, opt := range opts {
		opt(instance)
	}
	// 循环读取
	go instance.readLoop()
	// 心跳检测
	go instance.keepAlive()
	// 初始化返回
	if instance.initMsg != "" {
		logger.Debug("<== send welcome message ...")
		instance.initSendMsg()
	}
	return instance
}

func WithOnMessage(onMessage MessageHandle) Option {
	return func(c *conn) {
		c.onMessage = onMessage
	}
}

func WithTimeout(timeoutFunc TimeoutFunc) Option {
	return func(c *conn) {
		c.timeoutFunc = timeoutFunc
	}
}

func WithInitMsg(msg string) Option {
	return func(c *conn) {
		c.initMsg = msg
	}
}

func WithChangeDeviceIDHook(dh DeviceIDSetHook) Option {
	return func(c *conn) {
		c.onChangeDeviceId = dh
	}
}

func WithCheckDuration(t time.Duration) Option {
	return func(c *conn) {
		c.checkDuration = t
	}
}

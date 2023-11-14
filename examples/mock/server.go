package mock

import (
	x "X_IM"
	"X_IM/pkg/logger"
	"X_IM/pkg/naming"
	"X_IM/pkg/tcp"
	"X_IM/pkg/websocket"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type ServerDemo struct{}

type ServerHandler struct {
}

func (s *ServerDemo) Start(id, protocol, addr string) {
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	var srv x.Server
	service := &naming.DefaultService{
		ID:       id,
		Protocol: protocol,
	}
	if protocol == "ws" {
		srv = websocket.NewServer(addr, service)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr, service)
	}

	handler := &ServerHandler{}

	srv.SetReadWait(time.Minute)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

// Accept this connection
func (h *ServerHandler) Accept(conn x.Conn, timeout time.Duration) (string, x.Meta, error) {
	// 1. 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}
	logger.Info("in examples/mock/logic.go:Accept(): received an opcode:", frame.GetOpCode())
	// 2. 解析：数据包内容就是userId
	userID := string(frame.GetPayload())
	// 3. 鉴权：这里只是为了示例做一个fake验证，非空
	if userID == "" {
		return "", nil, errors.New("user id is invalid")
	}
	logger.Infof("in examples/mock/logic.go:Accept(): logined %s", userID)
	return userID, nil, nil
}

// Receive default listener
func (h *ServerHandler) Receive(ag x.Agent, payload []byte) {
	logger.Infof("srv received %s", string(payload))
	_ = ag.Push([]byte("ok"))
}

// Disconnect default listener
func (h *ServerHandler) Disconnect(id string) error {
	logger.Warnf("in examples/mock/logic.go:Disconnect(): disconnecter id: %s", id)
	return nil
}

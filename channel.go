package X_IM

import (
	"X_IM/logger"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ChannelImpl is a websocket implement of channel
type ChannelImpl struct {
	sync.Mutex
	id string
	Conn
	writeChan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	closed    *Event
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "tcp_channel",
		"id":     id,
	})
	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		writeChan: make(chan []byte, 5),
		writeWait: DefaultWriteWait,
		readWait:  DefaultReadWait,
		closed:    NewEvent(),
	}
	go func() {
		err := ch.writeLoop()
		if err != nil {
			log.Info(err)
		}
	}()
	return ch
}
func (ch *ChannelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}
	ch.writeWait = writeWait
}

func (ch *ChannelImpl) SetReadWait(readWait time.Duration) {
	if readWait == 0 {
		return
	}
	ch.readWait = readWait
}

func (ch *ChannelImpl) ID() string {
	return ch.id
}

// writeLoop 发送的消息直接通过writeChan发送给了一个独立的goroutine中writeLoop()执行
// 这样就使得Push变成了一个线程安全方法。
func (ch *ChannelImpl) writeLoop() error {
	for {
		select {
		case payload := <-ch.writeChan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			//批量写
			chanLen := len(ch.writeChan)
			for i := 0; i < chanLen; i++ {
				payload := <-ch.writeChan
				err := ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

func (ch *ChannelImpl) Push(payload []byte) error {
	// 通过原子操作保证了Channel的线程安全
	if ch.closed.HasFired() {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	// 异步写
	ch.writeChan <- payload
	return nil
}

// WriteFrame 重写conn的WriteFrame方法
// 增加了重置写超时的逻辑
func (ch *ChannelImpl) WriteFrame(op OpCode, payload []byte) error {
	_ = ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	return ch.Conn.WriteFrame(op, payload)
}

// ReadLoop 这是一个阻塞的方法，并且只允许被一个线程读取
// 因此我们直接在前面加了锁ch.Lock()，防止上层多次调用。
func (ch *ChannelImpl) ReadLoop(lst MessageListener) error {
	ch.Lock()
	defer ch.Unlock()
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "ReadLoop",
		"id":     ch.id,
	})
	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readWait))

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side closed the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Trace("recv a ping; resp with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		//Channel的生命周期是被通信层中的Server管理的
		//不希望其被上层消息处理器MessageListener直接操作，比如误调用Close()导致连接关闭。
		go lst.Receive(ch, payload)
	}
}

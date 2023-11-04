package X_IM

import (
	"X_IM/logger"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync/atomic"
	"time"
)

// ChannelImpl is a websocket implement of channel
type ChannelImpl struct {
	id string
	Conn
	meta      Meta
	writeChan chan []byte
	writeWait time.Duration
	readWait  time.Duration
	gpool     *ants.Pool
	state     int32
}

func NewChannel(id string, meta Meta, conn Conn, gpool *ants.Pool) Channel {
	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		meta:      meta,
		writeChan: make(chan []byte, 5),
		writeWait: DefaultWriteWait, //default value
		readWait:  DefaultReadWait,
		gpool:     gpool,
		state:     0,
	}
	go func() {
		err := ch.writeLoop()
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "ChannelImpl",
				"id":     id,
			}).Info(err)
		}
	}()
	return ch
}

func (ch *ChannelImpl) writeLoop() error {
	log := logger.WithFields(logger.Fields{
		"module": "ChannelImpl",
		"func":   "writeLoop",
		"id":     ch.id,
	})
	defer func() {
		log.Debugf("channel %s writeloop exited", ch.id)
	}()
	for payload := range ch.writeChan {
		err := ch.WriteFrame(OpBinary, payload)
		if err != nil {
			return err
		}
		chanLen := len(ch.writeChan)
		for i := 0; i < chanLen; i++ {
			payload = <-ch.writeChan
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
		}
		err = ch.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

// ID simpling logic
func (ch *ChannelImpl) ID() string { return ch.id }

// Push 异步写数据
func (ch *ChannelImpl) Push(payload []byte) error {
	if atomic.LoadInt32(&ch.state) != 1 {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	// 异步写
	ch.writeChan <- payload
	return nil
}

// Close 关闭连接
func (ch *ChannelImpl) Close() error {
	if !atomic.CompareAndSwapInt32(&ch.state, 1, 2) {
		return fmt.Errorf("channel has started")
	}
	close(ch.writeChan)
	return nil
}

// SetWriteWait 设置写超时
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
	ch.writeWait = readWait
}

func (ch *ChannelImpl) ReadLoop(lst MessageListener) error {
	// Perform an atomic compare-and-swap operation on ch.state. If the current value is 0, set it to 1.
	// If it's already 1, return an error indicating that the channel has already started.
	if !atomic.CompareAndSwapInt32(&ch.state, 0, 1) {
		return fmt.Errorf("channel has started")
	}

	// Create a new logger object with some fields filled out.
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "ReadLoop",
		"id":     ch.id,
	})

	// Start an infinite loop.
	for {
		// Set a read deadline for the channel.
		_ = ch.SetReadDeadline(time.Now().Add(ch.readWait))

		// Attempt to read a frame from the channel.
		frame, err := ch.ReadFrame()
		if err != nil {
			// If reading the frame failed, log the error and return it.
			log.Info(err)
			return err
		}
		if frame.GetOpCode() == OpClose {
			// If the received frame has an OpCode of OpClose, return an error indicating that the remote side closed the channel.
			return errors.New("remote side closed the channel")
		}
		if frame.GetOpCode() == OpPing {
			// If the received frame has an OpCode of OpPing, log that we received a ping and respond with a pong.
			log.Trace("recv a ping; resp with a pong")

			_ = ch.WriteFrame(OpPong, nil)
			_ = ch.Flush()
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			// If the payload is empty, skip to the next iteration of the loop.
			continue
		}
		err = ch.gpool.Submit(func() {
			// Submit a new task to the gpool (which is an instance of a goroutine pool).
			// This task calls lst.Receive with the channel and payload as parameters.
			lst.Receive(ch, payload)
		})
		if err != nil {
			// If submitting the task to the gpool failed, return an error.
			return err
		}
	}
}

func (ch *ChannelImpl) GetMeta() Meta { return ch.meta }

package pkt

import (
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/endian"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"strconv"
	"strings"
)

// LogicPkt 定义Gateway对外的Client消息结构
type LogicPkt struct {
	Header
	Body []byte `json:"body,omitempty"`
}

// HeaderOption 为了创建对象时参数可选，通过闭包实现选项模式
// Withxxx是默认的函数名称格式
type HeaderOption func(*Header)

func WithStatus(status Status) HeaderOption {
	return func(h *Header) {
		h.Status = status
	}
}

func WithSeq(seq uint32) HeaderOption {
	return func(h *Header) {
		h.Sequence = seq
	}
}

func WithChannel(channelID string) HeaderOption {
	return func(h *Header) {
		h.ChannelID = channelID
	}
}

func WithDest(dest string) HeaderOption {
	return func(h *Header) {
		h.Dest = dest
	}
}

// New an empty payload message
func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command

	for _, option := range options {
		option(&pkt.Header)
	}
	if pkt.Sequence == 0 {
		pkt.Sequence = common.Seq.Next()
	}
	return pkt
}

// NewFrom new a packet from a header
func NewFrom(header *Header) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Header = Header{
		Command:   header.Command,
		Sequence:  header.Sequence,
		ChannelID: header.ChannelID,
		Status:    header.Status,
		Dest:      header.Dest,
	}
	return pkt
}

// Decode read bytes to LogicPkt from a reader
func (p *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}
	// read body
	p.Body, err = endian.ReadBytes(r)
	if err != nil {
		return err
	}
	return nil
}

// Encode Header to writer.
// first use proto.Marshal
func (p *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}
	if err := endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, p.Body); err != nil {
		return err
	}
	return nil
}

// ReadBody val must be a pointer
func (p *LogicPkt) ReadBody(val proto.Message) error {
	return proto.Unmarshal(p.Body, val)
}

func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}
	p.Body, _ = proto.Marshal(val)
	return p
}

// StringBody return string body
func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

func (p *LogicPkt) String() string {
	return fmt.Sprintf("header:%v body:%dbits", &p.Header, len(p.Body))
}

// ServiceName 从消息头的command中获取服务名
func (h *Header) ServiceName() string {
	arr := strings.SplitN(h.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}
	return arr[0]
}

func (p *LogicPkt) AddMeta(m ...*Meta) {
	p.Meta = append(p.Meta, m...)
}

func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

// GetMeta extra value
func (p *LogicPkt) GetMeta(key string) (any, bool) {
	return FindMeta(p.Meta, key)
}

func FindMeta(meta []*Meta, key string) (any, bool) {
	for _, m := range meta {
		if m.Key == key {
			switch m.Type {
			case MetaType_int:
				v, _ := strconv.Atoi(m.Value)
				return v, true
			case MetaType_float:
				v, _ := strconv.ParseFloat(m.Value, 64)
				return v, true
			}
			return m.Value, true
		}
	}
	return nil, false
}

func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)
			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}
			p.Meta = p.Meta[:length-1]
		}
	}
}

package pkt

import (
	log "X_IM/pkg/logger"
	"X_IM/pkg/wire/common"
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type Packet interface {
	Decode(r io.Reader) error
	Encode(w io.Writer) error
}

func MustReadLogicPkt(r io.Reader) (*LogicPkt, error) {
	val, err := Read(r)
	if err != nil {
		return nil, err
	}
	if lp, ok := val.(*LogicPkt); ok {
		return lp, nil
	}
	return nil, fmt.Errorf("packet is not a logic packet")
}

//func MustReadBasicPkt(r io.Reader) (*BasicPkt, error) {
//	val, err := Read(r)
//	if err != nil {
//		return nil, err
//	}
//	if bp, ok := val.(*BasicPkt); ok {
//		return bp, nil
//	}
//	return nil, fmt.Errorf("packet is not a basic packet")
//}

// Read 从buf中读取magic code进行判断，随后根据magic code进行反序列化
func Read(r io.Reader) (any, error) {
	magic := common.Magic{}
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}
	switch magic {
	case common.MagicLogicPkt:
		fmt.Printf("magic code %s is correct\n", magic)
		p := new(LogicPkt)
		if err := p.Decode(r); err != nil {
			log.Warn("in pkt/packet.go:Read():decode failed.")
			return nil, err
		}

		return p, nil
	case common.MagicBasicPkt:
		p := new(BasicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("magic code %s is incorrect", magic)
	}
}

// Marshal 根据类型写入魔数到buf，随后将序列化后的字节数组写入其后
func Marshal(p Packet) []byte {
	buf := new(bytes.Buffer)
	kind := reflect.TypeOf(p).Elem()

	if kind.AssignableTo(reflect.TypeOf(LogicPkt{})) {
		_, _ = buf.Write(common.MagicLogicPkt[:])
	} else if kind.AssignableTo(reflect.TypeOf(BasicPkt{})) {
		_, _ = buf.Write(common.MagicBasicPkt[:])
	}
	_ = p.Encode(buf)
	return buf.Bytes()
}

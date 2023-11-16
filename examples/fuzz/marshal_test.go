package fuzz

import (
	"X_IM/pkg/wire/common"
	"X_IM/pkg/wire/pkt"
	"bytes"
	"testing"
)

func FuzzMarshal(f *testing.F) {
	f.Add([]byte("test data")) // 添加一些初始的测试数据

	f.Fuzz(func(t *testing.T, data []byte) {
		// 1. 读取登录包
		buf := bytes.NewBuffer(data)
		req, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			t.Skip("in Accept(): ", err)
		}

		// 2. 必须是登录包
		if req.Command != common.CommandLoginSignIn {
			resp := pkt.NewFrom(&req.Header)
			resp.Status = pkt.Status_InvalidCommand
			t.Skip("is an invalid command")
		}

		// 3. unmarshal Body
		var login pkt.LoginReq
		err = req.ReadBody(&login)
		if err != nil {
			t.Error(err)
		}

		// 4. 修改并重新序列化
		req.Command = common.CommandLoginSignOut
		buf.Reset()
		err = req.Encode(buf)
		if err != nil {
			t.Error(err)
		}

		// 5. 重新读取并检查
		req2, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			t.Error(err)
		}
		if req2.Command != common.CommandLoginSignOut {
			t.Error("Command was not modified correctly")
		}
	})
}

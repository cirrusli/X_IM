package restful

import (
	"X_IM/pkg/logger"
	"X_IM/wire/rpc"
	"fmt"
	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
	"time"
)

type Message interface {
	InsertUser(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error)
	InsertGroup(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error)
	SetACK(app string, req *rpc.AckMessageReq) error
	GetMessageIndex(app string, req *rpc.GetOfflineMessageIndexReq) (*rpc.GetOfflineMessageIndexResp, error)
	GetMessageContent(app string, req *rpc.GetOfflineMessageContentReq) (*rpc.GetOfflineMessageContentResp, error)
}
type MessageHTTP struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
}

func NewMessageService(url string) Message {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	return &MessageHTTP{
		url: url,
		cli: cli,
	}
}

func NewMessageServiceWithSRV(scheme string, srv *resty.SRVRecord) Message {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme("http")

	return &MessageHTTP{
		url: "",
		cli: cli,
		srv: srv,
	}
}

func (m *MessageHTTP) InsertUser(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/user", m.url, app)
	t1 := time.Now()

	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHTTP.InsertUser response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("MessageHTTP.InsertUser cost %v resp: %v", time.Since(t1), &resp)
	return &resp, nil
}

func (m *MessageHTTP) InsertGroup(app string, req *rpc.InsertMessageReq) (*rpc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/group", m.url, app)
	t1 := time.Now()
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHTTP.InsertGroup response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("MessageHTTP.InsertGroup cost %v resp: %v", time.Since(t1), &resp)
	return &resp, nil
}

func (m *MessageHTTP) SetACK(app string, req *rpc.AckMessageReq) error {
	path := fmt.Sprintf("%s/api/%s/message/ack", m.url, app)
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("MessageHTTP.SetACK response.StatusCode() = %d, want 200", response.StatusCode())
	}
	return nil
}

func (m *MessageHTTP) GetMessageIndex(app string, req *rpc.GetOfflineMessageIndexReq) (*rpc.GetOfflineMessageIndexResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/index", m.url, app)
	body, _ := proto.Marshal(req)

	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHTTP.GetMessageIndex response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.GetOfflineMessageIndexResp
	_ = proto.Unmarshal(response.Body(), &resp)
	return &resp, nil
}

func (m *MessageHTTP) GetMessageContent(app string, req *rpc.GetOfflineMessageContentReq) (*rpc.GetOfflineMessageContentResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/content", m.url, app)
	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHTTP.GetMessageContent response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.GetOfflineMessageContentResp
	_ = proto.Unmarshal(response.Body(), &resp)
	return &resp, nil
}

func (m *MessageHTTP) Req() *resty.Request {
	if m.srv == nil {
		return m.cli.R()
	}
	return m.cli.R().SetSRV(m.srv)
}

package client

import (
	"X_IM/pkg/logger"
	"X_IM/pkg/wire/rpc"
	"fmt"
	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
	"time"
)

type Group interface {
	Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error)
	Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error)
	Join(app string, req *rpc.JoinGroupReq) error
	Quit(app string, req *rpc.QuitGroupReq) error
	Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error)
}
type GroupHTTP struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
}

func NewGroupService(url string) Group {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme("http")
	return &GroupHTTP{
		url: url,
		cli: cli,
	}
}

func NewGroupServiceWithSRV(scheme string, srv *resty.SRVRecord) Group {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme("http")

	return &GroupHTTP{
		url: "",
		cli: cli,
		srv: srv,
	}
}

func (g *GroupHTTP) Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group", g.url, app)

	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHTTP.Create response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.CreateGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHTTP.Create resp: %v", &resp)
	return &resp, nil
}

func (g *GroupHTTP) Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/members/%s", g.url, app, req.GroupID)

	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHTTP.Members response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.GroupMembersResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHTTP.Members resp: %v", &resp)
	return &resp, nil
}

func (g *GroupHTTP) Join(app string, req *rpc.JoinGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHTTP.Join response.StatusCode() = %d, want 200", response.StatusCode())
	}
	return nil
}

func (g *GroupHTTP) Quit(app string, req *rpc.QuitGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Req().SetBody(body).Delete(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHTTP.Quit response.StatusCode() = %d, want 200", response.StatusCode())
	}
	return nil
}

func (g *GroupHTTP) Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/%s", g.url, app, req.GroupID)
	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHTTP.Detail response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.GetGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHTTP.Detail resp: %v", &resp)
	return &resp, nil
}

// Req 注入SRV配置
// 默认的DNS域名解析功能是很弱的，只返回一个IP
// 对内部的服务之间的调用来说，至少要给出服务的访问端口
// 因此需要SRV record
func (g *GroupHTTP) Req() *resty.Request {
	if g.srv == nil {
		return g.cli.R()
	}
	return g.cli.R().SetSRV(g.srv)
}

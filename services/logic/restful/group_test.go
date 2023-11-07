package restful

import (
	"X_IM/wire/rpc"
	"testing"

	"github.com/stretchr/testify/assert"
)

const app = "x_im_test"

var groupService = NewGroupService("http://localhost:8080")

func TestGroupService(t *testing.T) {

	resp, err := groupService.Create(app, &rpc.CreateGroupReq{
		Name:    "test",
		Owner:   "test1",
		Members: []string{"test1", "test2"},
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.GroupID)
	t.Log(resp.GroupID)

	mresp, err := groupService.Members(app, &rpc.GroupMembersReq{
		GroupID: resp.GroupID,
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(mresp.Users))
	assert.Equal(t, "test1", mresp.Users[0].Account)
	assert.Equal(t, "test2", mresp.Users[1].Account)

	err = groupService.Join(app, &rpc.JoinGroupReq{
		Account: "test3",
		GroupID: resp.GroupID,
	})
	assert.Nil(t, err)

	mresp, err = groupService.Members(app, &rpc.GroupMembersReq{
		GroupID: resp.GroupID,
	})
	assert.Nil(t, err)

	assert.Equal(t, 3, len(mresp.Users))
	assert.Equal(t, "test3", mresp.Users[2].Account)
	assert.Equal(t, "test2", mresp.Users[1].Account)

	err = groupService.Quit(app, &rpc.QuitGroupReq{
		Account: "test2",
		GroupID: resp.GroupID,
	})
	assert.Nil(t, err)

	mresp, err = groupService.Members(app, &rpc.GroupMembersReq{
		GroupID: resp.GroupID,
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(mresp.Users))
	assert.Equal(t, "test1", mresp.Users[0].Account)
	assert.Equal(t, "test3", mresp.Users[1].Account)
}

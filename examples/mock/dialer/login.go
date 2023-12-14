package dialer

import (
	"X_IM/pkg/token"
	"X_IM/pkg/websocket"
	"X_IM/pkg/x"
)

func Login(wsurl, account string, appSecrets ...string) (x.Client, error) {
	cli := websocket.NewClient(account, "unit_test", websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
		//使用传入的第一个appSecret
		secret = appSecrets[0]
	}
	cli.SetDialer(&ClientDialer{
		AppSecret: secret,
	})
	err := cli.Connect(wsurl)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

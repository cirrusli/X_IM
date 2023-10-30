package dialer

import (
	x "X_IM"
	"X_IM/websocket"
	"X_IM/wire/token"
)

func Login(wsurl, account string, appSecrets ...string) (x.Client, error) {
	cli := websocket.NewClient(account, "unit_test", websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
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

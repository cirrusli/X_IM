package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	cli := resty.New()
	cli.SetHeader("Content-Type", "application/json")

	domains := make(map[string]int)
	for i := 0; i < 10; i++ {
		//use ksuid to generate a random token
		url := fmt.Sprintf("http://localhost:8100/api/lookup/%s", ksuid.New().String())

		var res LookUpResp //定义一个结构体，用于接收返回的json数据
		resp, err := cli.R().SetResult(&res).Get(url)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Nil(t, err)
		if len(res.Domains) > 0 {
			domain := res.Domains[0]
			domains[domain]++
		}
	}
	for domain, hit := range domains {
		fmt.Printf("domain: %s ;hit count: %d\n", domain, hit)
	}
}

package conf

import (
	"encoding/json"
	"github.com/bytedance/sonic"
	"os"
	"testing"
	"time"
)

//如果没有时间消耗，是因为测试用的json太小，不能用来分析性能差距……
//平均每次操作耗时都是0.0005ns左右

func BenchmarkSonicUnmarshal(b *testing.B) {
	// 读取JSON文件
	data, err := os.ReadFile("../route.json")
	if err != nil {
		b.Fatalf("Failed to read JSON file: %v", err)
	}

	var conf struct {
		RouteBy   string `json:"route_by,omitempty"`
		Zones     []Zone `json:"zones,omitempty"`
		Whitelist []struct {
			Key   string `json:"key,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"whitelist,omitempty"`
	}

	// 使用sonic.Unmarshal解析JSON
	start := time.Now()
	err = sonic.Unmarshal(data, &conf)
	if err != nil {
		b.Fatalf("Failed to unmarshal JSON with sonic: %v", err)
	}
	elapsed := time.Since(start)

	b.Logf("Sonic unmarshal took %s", elapsed)
}

func BenchmarkJsonUnmarshal(b *testing.B) {
	// 读取JSON文件
	data, err := os.ReadFile("../route.json")
	if err != nil {
		b.Fatalf("Failed to read JSON file: %v", err)
	}

	var conf struct {
		RouteBy   string `json:"route_by,omitempty"`
		Zones     []Zone `json:"zones,omitempty"`
		Whitelist []struct {
			Key   string `json:"key,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"whitelist,omitempty"`
	}

	// 使用json.Unmarshal解析JSON
	start := time.Now()
	err = json.Unmarshal(data, &conf)
	if err != nil {
		b.Fatalf("Failed to unmarshal JSON with json: %v", err)
	}
	elapsed := time.Since(start)

	b.Logf("Json unmarshal took %s", elapsed)
}

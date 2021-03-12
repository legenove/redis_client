package rediscore

import (
	"context"
	"fmt"
	"testing"

	"github.com/legenove/cocore"
)

func init() {
	cocore.InitApp(true, "", cocore.AppParam{
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/redis_client/conf",
		},
	})
}

func TestGetRedisClient(t *testing.T) {
	client, err := GetRedisClient("default_redis")
	if err != nil {
		fmt.Println(err)
		return
	}
	x := client.Ping(context.Background())
	fmt.Println(x.Val())
	fmt.Println(x.Err())
	fmt.Println(x.Result())
}

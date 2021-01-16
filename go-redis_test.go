package redis_client

import (
	"fmt"
	"github.com/legenove/cocore"
	"testing"
)

func init() {
	cocore.InitApp(true, "", cocore.ConfigParam{
		Type:      cocore.TYPE_CONFIG_FILE,
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
	fmt.Println(client.Ping().Result())
}

package redis_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/legenove/utils"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/legenove/cocore"
	"github.com/legenove/easyconfig/ifacer"
)

type closer interface {
	Close() error
}

type mangers struct {
	clients  map[string]*redis.Client
	clusters map[string]*redis.ClusterClient
	sync.Mutex
}

type closeManagers struct {
	cli closer
	rat int64
}

var redisConf ifacer.Configer
var Manager = &mangers{clients: make(map[string]*redis.Client), clusters: make(map[string]*redis.ClusterClient)}
var pendclose = map[string]closeManagers{}
var redisSettings = make(map[string]*RedisSetting)

//func init() {
//	keeper.SetKeeper(random.UuidV5(), closeRedis, time.Second*600, false)
//}

func loadRedisManager(rsettings []*RedisSetting, b ...bool) {
	preload := true
	if len(b) > 0 {
		preload = b[0]
	}
	for _, rsetting := range rsettings {
		if rsetting.Type == RedisTypeMaster || rsetting.Type == RedisTypeSlaver {
			if _, ok := Manager.clients[rsetting.RouterName]; preload && !ok {
				Manager.clients[rsetting.RouterName] = newRedisClient(rsetting)
			}
		} else if rsetting.Type == RedisTypeCluster {
			if _, ok := Manager.clusters[rsetting.RouterName]; preload && !ok {
				// TODO
			}

		} else {
			panic("redis conf type not support :" + rsetting.Type)
		}
		if _, ok := redisSettings[rsetting.RouterName]; !ok {
			redisSettings[rsetting.RouterName] = rsetting
		}
	}
}

func removeRedis() {
	Manager.Lock()
	defer Manager.Unlock()
	nowAt := time.Now()
	nowS := nowAt.Format("20060102150405.999999")
	nowT := nowAt.Unix()
	for k, v := range Manager.clients {
		pendclose[utils.ConcatenateStrings(k, nowS, "_client")] = closeManagers{
			cli: v,
			rat: nowT,
		}
	}
	Manager.clients = make(map[string]*redis.Client)
	for k, v := range Manager.clusters {
		pendclose[utils.ConcatenateStrings(k, nowS, "_cluster")] = closeManagers{
			cli: v,
			rat: nowT,
		}
	}
	Manager.clusters = make(map[string]*redis.ClusterClient)
	redisSettings = make(map[string]*RedisSetting)
}

func closeRedis() error {
	Manager.Lock()
	defer Manager.Unlock()
	nowAt := time.Now().Unix()
	for k, v := range pendclose {
		if nowAt-v.rat > 600 {
			// 十分钟后自动关闭，防止链接泄漏
			v.cli.Close()
			delete(pendclose, k)
		}
	}
	return nil
}

func getRedisConf(key string) (*RedisSetting, error) {
	s, ok := redisSettings[key]
	if ok {
		return s, nil
	}
	if redisConf == nil {
		err := newRedisConfig()
		if err != nil {
			return nil, err
		}
		if redisConf == nil {
			return nil, fmt.Errorf("redis conf not setting")
		}
	}
	var setting RedisSetting
	err := redisConf.UnmarshalKey(key, &setting)
	if err != nil {
		return nil, fmt.Errorf("Invalid redis conf:%s; err:%s", key, err.Error())
	}
	return &setting, nil
}

func newRedisConfig() error {
	var err error
	redisFileName := cocore.App.GetStringConfig("redis_conf", "redis.yaml")
	redisFileType := cocore.App.GetStringConfig("redis_conf_type", "yaml")
	redisConf, err = cocore.Conf.Instance(redisFileName, redisFileType, nil)
	go listenOnRedisChange(redisConf)
	return err
}

func listenOnRedisChange(v ifacer.Configer) {
	if v != nil {
		<-v.OnChangeChan()
		for {
			select {
			case <-v.OnChangeChan():
				removeRedis()
			}
		}
	}
}

func GetRedisClient(key string) (*redis.Client, error) {
	return Manager.GetRedisClient(key)
}

// Todo 先预留，没有测试过，setting可能和单点redis不同。
func GetRedisCluster(key string) (*redis.ClusterClient, error) {
	return Manager.GetRedisClusterClient(key)
}

func (m *mangers) GetRedisClient(key string) (*redis.Client, error) {
	m.Lock()
	redisSetting, err := getRedisConf(key)
	m.Unlock()
	if err != nil {
		return nil, err
	}
	m.Lock()
	client, ok := m.clients[key]
	if !ok {
		loadRedisManager([]*RedisSetting{redisSetting})
		client, ok = m.clients[key]
		if !ok {
			m.Unlock()
			return nil, errors.New(fmt.Sprintf("redis client can't be created, url: %s", redisSetting.Url))
		}
	}
	m.Unlock()
	return client, nil
}

func (m *mangers) GetRedisClusterClient(key string) (*redis.ClusterClient, error) {
	m.Lock()
	redisSetting, err := getRedisConf(key)
	m.Unlock()
	if err != nil {
		return nil, err
	}
	if redisSetting.Type != RedisTypeCluster {
		return nil, errors.New(fmt.Sprintf("%s : redis client type not cluster: %s",
			redisSetting.RouterName, redisSetting.Type))
	}

	m.Lock()
	defer m.Unlock()
	client, ok := m.clusters[key]
	if !ok {
		loadRedisManager([]*RedisSetting{redisSetting})
		client, ok = m.clusters[key]
		if !ok {
			return nil, errors.New(fmt.Sprintf("%s : redis client can't be created, url: %s",
				redisSetting.RouterName, redisSetting.Url))
		}
	}
	return client, nil
}

func newRedisClient(setting *RedisSetting) *redis.Client {
	opt := &redis.Options{
		Addr:               setting.GetAddr(),
		Password:           setting.GetPassword(),
		PoolSize:           setting.GetPoolSize(),
		MinIdleConns:       setting.GetMinIdleConns(),
		DialTimeout:        setting.GetDialTimeout(),
		ReadTimeout:        setting.GetReadTimeout(),
		WriteTimeout:       setting.GetWriteTimeout(),
		IdleTimeout:        setting.GetIdleTimeout(),
		IdleCheckFrequency: setting.GetIdleCheckFrequency(),
		OnConnect: func(ctx context.Context, conn *redis.Conn) error {
			_, err := conn.Ping(ctx).Result()
			return err
		},
	}
	return redis.NewClient(opt)
}

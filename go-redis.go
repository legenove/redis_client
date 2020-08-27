package redis_client

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/legenove/cocore"
	"github.com/legenove/viper_conf"
	"sync"
)

type mangers struct {
	clients  map[string]*redis.Client
	clusters map[string]*redis.ClusterClient
	sync.Mutex
}

var redisConf *viper_conf.ViperConf
var Manager = &mangers{clients: make(map[string]*redis.Client), clusters: make(map[string]*redis.ClusterClient)}
var redisSettings = make(map[string]*RedisSetting)

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
	Manager = &mangers{clients: make(map[string]*redis.Client), clusters: make(map[string]*redis.ClusterClient)}
	redisSettings = make(map[string]*RedisSetting)
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
	err := redisConf.GetConf().UnmarshalKey(key, &setting)
	if err != nil {
		return nil, fmt.Errorf("Invalid redis conf:%s; err:%s", key, err.Error())
	}
	return &setting, nil
}

func newRedisConfig() error {
	var err error
	redisFileName := cocore.App.GetStringConfig("redis_conf", "redis.toml")
	redisConf, err = cocore.Conf.Instance(redisFileName, nil)
	go listenOnRedisChange(redisConf)
	return err
}

func listenOnRedisChange(v *viper_conf.ViperConf) {
	if v != nil {
		<- v.OnChange
		for {
			select {
			case <- v.OnChange:
				removeRedis()
			}
		}
	}
}

func GetRedisClient(key string) (*redis.Client, error) {
	return Manager.GetRedisClient(key)
}

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
		OnConnect: func(conn *redis.Conn) error {
			_, err := conn.Ping().Result()
			return err
		},
	}
	return redis.NewClient(opt)
}

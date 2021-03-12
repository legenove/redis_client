package rediscore

import "time"

const (
	RedisTypeMaster  = "master"
	RedisTypeSlaver  = "slaver"
	RedisTypeCluster = "cluster"
)

type RedisSetting struct {
	RouterName         string
	Type               string
	Url                string // host:port
	Password           string // password
	DB                 int    // 数据库选择
	PoolSize           int    // 链接池最大数量,  go-redis 默认 10
	MinIdleConns       int    // 链接池最小存活链接数量， 默认无
	DialTimeout        int    // 创建链接超时,  go-redis 默认 5 s
	ReadTimeout        int    // 读超时，毫秒, go-redis 默认 3000 毫秒
	WriteTimeout       int    // 写超时，毫秒, go-redis 默认 ReadTimeout
	IdleTimeout        int    // 最后使用的空闲时间，后重新进行链接, go-redis 默认 5min
	IdleCheckFrequency int    // 默认检测时间, go-redis 默认 1min
	//Username           string // for redis 6.0
}

func (s *RedisSetting) GetReadOnly() bool {
	if s.Type == RedisTypeSlaver {
		return true
	}
	return false
}

func (s *RedisSetting) GetAddr() string {
	return s.Url
}

//func (s *RedisSetting) GetUserName() string {
//	return s.Username
//}

func (s *RedisSetting) GetPassword() string {
	return s.Password
}

func (s *RedisSetting) GetPoolSize() int {
	if s.PoolSize <= 0 {
		s.PoolSize = 0
	}
	return s.PoolSize
}

func (s *RedisSetting) GetMinIdleConns() int {
	if s.MinIdleConns <= 0 {
		s.MinIdleConns = 0
	}
	return s.MinIdleConns
}

func (s *RedisSetting) GetDialTimeout() time.Duration {
	if s.DialTimeout <= 0 {
		s.DialTimeout = 0
	}
	return time.Duration(s.DialTimeout) * time.Second
}

func (s *RedisSetting) GetReadTimeout() time.Duration {
	if s.ReadTimeout <= 0 {
		s.ReadTimeout = 0
	} else if s.ReadTimeout <= 200 {
		s.ReadTimeout = 200
	}
	return time.Duration(s.ReadTimeout) * time.Millisecond
}

func (s *RedisSetting) GetWriteTimeout() time.Duration {
	if s.WriteTimeout <= 0 {
		return s.GetReadTimeout()
	} else if s.WriteTimeout <= 200 {
		s.WriteTimeout = 200
	}
	return time.Duration(s.WriteTimeout) * time.Millisecond
}

func (s *RedisSetting) GetIdleTimeout() time.Duration {
	if s.IdleTimeout <= 0 {
		s.IdleTimeout = 0
	}
	return time.Duration(s.IdleTimeout) * time.Second
}

func (s *RedisSetting) GetIdleCheckFrequency() time.Duration {
	if s.IdleCheckFrequency <= 0 {
		s.IdleCheckFrequency = 0
	}
	return time.Duration(s.IdleCheckFrequency) * time.Second
}

package lib

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

//生成链接池
func InitRedisPool(path string) error {
	RedisConfMap := &RedisMapConf{}
	err := ParseConfig(path, RedisConfMap)
	if err != nil {
		return err
	}
	ConfRedisMap = RedisConfMap
	if len(ConfRedisMap.List) == 0 {
		fmt.Printf("[INFO] %s%s\n", time.Now().Format(TimeFormat), " empty redis config.")
	}

	RedisMapPool = map[string]*redis.Pool{}

	if ConfRedisMap != nil && ConfRedisMap.List != nil {
		for confName, cfg := range ConfRedisMap.List {
			if cfg.ConnTimeout == 0 {
				cfg.ConnTimeout = 50
			}
			if cfg.ReadTimeout == 0 {
				cfg.ReadTimeout = 100
			}
			if cfg.WriteTimeout == 0 {
				cfg.WriteTimeout = 100
			}
			redispool := &redis.Pool{
				MaxIdle:     cfg.MaxIdle,
				MaxActive:   cfg.MaxActive,
				IdleTimeout: 240 * time.Second,
				Wait:        true,
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial(
						"tcp",
						cfg.ProxyList[0],
						redis.DialPassword(cfg.ProxyList[1]),
						redis.DialConnectTimeout(time.Duration(cfg.ConnTimeout)*time.Millisecond),
						redis.DialReadTimeout(time.Duration(cfg.ReadTimeout)*time.Millisecond),
						redis.DialWriteTimeout(time.Duration(cfg.WriteTimeout)*time.Millisecond))
					if err != nil {
						return nil, err
					}
					return c, nil
				},
			}
			RedisMapPool[confName] = redispool
		}
		if defaultpool, err := GetRedisPool("default"); err == nil {
			RedisDefaultPool = defaultpool
		}
	}
	return nil
}

//获取链接池
func GetRedisPool(name string) (*redis.Pool, error) {
	if redispool, ok := RedisMapPool[name]; ok {
		return redispool, nil
	}
	return nil, errors.New("get pool error")
}

//从连接池里获取一个连接
func RedisConnFactory(name string) (redis.Conn, error) {
	var (
		pool *redis.Pool
		err  error
	)
	if pool, err = GetRedisPool(name); err != nil {
		return nil, err
	}
	return pool.Get(), nil
}

//关闭一个链接
func RedisConnClose(trace *TraceContext, conn redis.Conn) {
	startExecTime := time.Now()
	if err := conn.Close(); err != nil {
		endExecTime := time.Now()
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"err":       errors.New("RedisConnCloseError"),
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return
}

func RedisLogDo(trace *TraceContext, c redis.Conn, commandName string, args ...interface{}) (interface{}, error) {
	startExecTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}

//通过配置 执行redis
func RedisConfDo(trace *TraceContext, name string, commandName string, args ...interface{}) (interface{}, error) {
	c, err := RedisConnFactory(name)
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method": commandName,
			"err":    errors.New("RedisConnFactory_error:" + name),
			"bind":   args,
		})
		return nil, err
	}
	defer RedisConnClose(trace, c)

	startExecTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}

package cache_storage

import (
	"auth_service_template/logger"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	ErrStorageTypeError = errors.New("Storage type error")
)

type TimeStorage struct {
	conn interface{}
	log  *logger.Log
}

func NewTimeStorage(glob_env *map[string]string, logs *logger.Log) *TimeStorage {
	storage := TimeStorage{
		log: logs,
	}
	var dialKeepAlive time.Duration
	if i32, err := strconv.Atoi((*glob_env)["STORAGE_KEEP_ALIVE"]); err == nil {
		dialKeepAlive = time.Duration(i32) * time.Second
	}

	if (*glob_env)["STORAGE_DRIVER"] == "redis" {
		c, err := redis.Dial("tcp",
			fmt.Sprintf(`%s:%s`, (*glob_env)["STORAGE_ADDR"], (*glob_env)["STORAGE_PORT"]),
			// Read timeout on server should be greater than ping period.
			redis.DialReadTimeout(time.Minute+10*time.Second),
			redis.DialWriteTimeout(10*time.Second),
			redis.DialKeepAlive(dialKeepAlive),
			redis.DialPassword((*glob_env)["STORAGE_PASSWORD"]),
		)
		if err != nil {
			logs.Println(logger.NewError(
				"redis",
				fmt.Sprintf("Could not connect to Redis server: %v", err),
				logger.FATAL,
			))
		} else {
			logs.Println(logger.NewError(
				"redis",
				fmt.Sprintf("Redis server connection success"),
				logger.INFO,
			))
		}
		storage.conn = c
	}
	return &storage
}

func (s *TimeStorage) Close() {
	switch c := s.conn.(type) {
	case redis.Conn:
		c.Close()
	}
}

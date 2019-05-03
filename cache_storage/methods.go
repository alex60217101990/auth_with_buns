package cache_storage

import (
	"auth_service_template/logger"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func (s *TimeStorage) Put(key string, data string, params map[string]interface{}) error {
	switch c := s.conn.(type) {
	case redis.Conn:
		var duration string
		for k, v := range params {
			if k == "duration" {
				duration = v.(string)
			}
		}
		if len(duration) > 0 {
			_, err := c.Do("SETEX", data, duration, key)
			if err != nil {
				// If there is an error in setting the cache, return an internal server error
				s.log.Println(logger.NewError(
					"redis",
					fmt.Sprintf("SETEX error: [%s]", err.Error()),
					logger.ERROR,
				))
				return err
			}
		} else {
			_, err := c.Do("SET", data, key)
			if err != nil {
				// If there is an error in setting the cache, return an internal server error
				s.log.Println(logger.NewError(
					"redis",
					fmt.Sprintf("SET error: [%s]", err.Error()),
					logger.ERROR,
				))
				return err
			}
		}
		return nil
	}
	return ErrStorageTypeError
}

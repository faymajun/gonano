package redisoperate

import (
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/faymajun/gonano/redis"

	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

func SetValue(key, field string, val interface{}) error {
	if _, ok := val.(msgp.Marshaler); ok {
		bin, err := val.(msgp.Marshaler).MarshalMsg(nil)
		if err != nil {
			return err
		}
		val = bin
	}

	redisClient := redis.LocalRedis()
	if redisClient == nil {
		logrus.Errorln("set redis error, errors is nil")
	}
	if err := redisClient.HSet(key, field, val).Err(); err != nil {
	}

	return nil
}

// 这里切记。fields 必须是新开的，因为要放在其他go使用
func MSetValue(key string, fields map[string]interface{}) error {
	for key, val := range fields {
		if _, ok := val.(msgp.Marshaler); ok {
			bin, err := val.(msgp.Marshaler).MarshalMsg(nil)
			if err != nil {
				return err
			}
			fields[key] = bin
		}
	}
	if err := redis.LocalRedis().HMSet(key, fields).Err(); err != nil {
	}

	return nil
}

func DelField(key string, field ...string) error {
	if err := redis.LocalRedis().HDel(key, field...).Err(); err != nil {

	}

	return nil
}

func IncrValue(key, field string, incr int64) error {
	if err := redis.LocalRedis().HIncrBy(key, field, incr).Err(); err != nil {
	}

	return nil
}

func GetString(key, field string) (string, error) {
	cmd := redis.LocalRedis().HGet(key, field)
	if cmd != nil {
		return cmd.Val(), nil
	}

	return "", errors.Errorf("get redis error, key=%s, field=%s", key, field)
}

func GetInt32(key, field string) (int32, error) {
	val, err := GetString(key, field)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(num), nil
}

func GetInt64(key, field string) (int64, error) {
	val, err := GetString(key, field)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func GetFloat32(key, field string) (float32, error) {
	val, err := GetString(key, field)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}

	return float32(num), nil
}

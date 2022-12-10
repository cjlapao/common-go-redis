package redisclient

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var globalRedisClient *RedisClient

type RedisConnectionString struct {
	Address  string
	Port     int
	Password string
	Database int
}

type RedisClient struct {
	context             context.Context
	client              *redis.Client
	rawConnectionString string
	connectionString    RedisConnectionString
}

func New(connString string) *RedisClient {
	result := RedisClient{
		context:             context.Background(),
		rawConnectionString: connString,
	}

	c := result.parseConnectionString(connString)
	result.connectionString = c

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Address, strconv.Itoa(c.Port)),
		Password: c.Password,
		DB:       c.Database,
	})
	result.client = rdb

	globalRedisClient = &result
	return globalRedisClient
}

func Get(connString string) *RedisClient {
	if globalRedisClient != nil {
		if globalRedisClient.rawConnectionString != connString {
			return New(connString)
		}

		return globalRedisClient
	}

	return New(connString)
}

func (cli RedisClient) Ping() error {
	pingCmd := cli.client.Ping(cli.context)
	_, err := pingCmd.Result()
	return err
}

func (cli RedisClient) Close() error {
	return cli.client.Close()
}

func (cli RedisClient) SetString(key string, value interface{}) error {
	return cli.SetExpiringString(key, value, 0)
}

func (cli RedisClient) SetExpiringString(key string, value interface{}, expireIn time.Duration) error {
	cmd := cli.client.Set(cli.context, key, value, expireIn)

	return cmd.Err()
}

func (cli RedisClient) GetStringKey(key string) (string, error) {
	cmd := cli.client.Get(cli.context, key)

	result, err := cmd.Result()
	if err == redis.Nil {
		return "", fmt.Errorf("the key does not exist")
	} else if err != nil {
		return "", err
	}

	return result, nil
}

func (cli RedisClient) GetAllKeys(prefix string) ([]string, error) {
	keys := make([]string, 0)

	iter := cli.client.Scan(cli.context, 0, fmt.Sprintf("%s*", prefix), 0).Iterator()
	for iter.Next(cli.context) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (cli RedisClient) AddToList(key string, values ...interface{}) error {
	cmd := cli.client.LPush(cli.context, key, values...)

	_, err := cmd.Result()

	return err
}

func (cli RedisClient) PopQueueList(key string) (string, error) {
	cmd := cli.client.RPop(cli.context, key)

	string, err := cmd.Result()

	return string, err
}

func (cli RedisClient) PopStackList(key string) (string, error) {
	cmd := cli.client.LPop(cli.context, key)

	val, err := cmd.Result()

	return val, err
}

func (cli RedisClient) GetListCount(key string) (int64, error) {
	cmd := cli.client.LLen(cli.context, key)

	count, err := cmd.Result()

	return count, err
}

func (cli RedisClient) TrimList(key string, from int64, to int64) error {
	cmd := cli.client.LTrim(cli.context, key, from, to)

	_, err := cmd.Result()

	return err
}

func (cli RedisClient) Delete(keys ...string) (int64, error) {
	cmd := cli.client.Del(cli.context, keys...)

	count, err := cmd.Result()

	return count, err
}

func (cli RedisClient) parseConnectionString(connString string) RedisConnectionString {
	result := RedisConnectionString{
		Port: 6379,
	}
	serverDbParts := strings.Split(connString, "/")
	if len(serverDbParts) == 1 {
		result.Address = connString
		result.Database = 0
	} else {
		result.Address = serverDbParts[0]
		if db, err := strconv.Atoi(serverDbParts[1]); err == nil {
			result.Database = db
		}
	}

	serverUserParts := strings.Split(result.Address, "@")
	if len(serverUserParts) > 1 {
		result.Address = serverUserParts[1]
		result.Password = serverUserParts[0]
	}

	serverPortParts := strings.Split(result.Address, ":")
	if len(serverPortParts) > 1 {
		result.Address = serverPortParts[0]
		if port, err := strconv.Atoi(serverPortParts[1]); err == nil {
			result.Port = port
		}
	}

	return result
}

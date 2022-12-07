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

func (cli RedisClient) Ping() error {
	pingCmd := cli.client.Ping(cli.context)
	_, err := pingCmd.Result()
	return err
}

func (cli RedisClient) Close() error {
	return cli.client.Close()
}

func (cli RedisClient) Set(key string, value interface{}) error {
	return cli.SetExpiring(key, value, 0)
}

func (cli RedisClient) SetExpiring(key string, value interface{}, expireIn time.Duration) error {
	cmd := cli.client.Set(cli.context, key, value, expireIn)

	return cmd.Err()
}

func (cli RedisClient) Get(key string) (string, error) {
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

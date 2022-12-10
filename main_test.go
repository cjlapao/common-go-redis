package redisclient

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/stretchr/testify/assert"
)

var redisServer *miniredis.Miniredis

func TestRedisClient_parseConnectionString(t *testing.T) {
	type args struct {
		connString string
	}
	tests := []struct {
		name string
		cli  RedisClient
		args args
		want RedisConnectionString
	}{
		{
			name: "full string",
			cli:  RedisClient{},
			args: args{
				connString: "password@localhost/0",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "password",
				Port:     6379,
				Database: 0,
			},
		},
		{
			name: "host",
			cli:  RedisClient{},
			args: args{
				connString: "localhost",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "",
				Port:     6379,
				Database: 0,
			},
		},
		{
			name: "host with db",
			cli:  RedisClient{},
			args: args{
				connString: "localhost/3",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "",
				Port:     6379,
				Database: 3,
			},
		},
		{
			name: "full string with db and port",
			cli:  RedisClient{},
			args: args{
				connString: "password@localhost:5500/3",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "password",
				Port:     5500,
				Database: 3,
			},
		},
		{
			name: "host and password",
			cli:  RedisClient{},
			args: args{
				connString: "password@localhost:5500",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "password",
				Port:     5500,
				Database: 0,
			},
		},
		{
			name: "host and password",
			cli:  RedisClient{},
			args: args{
				connString: "password@localhost:abc",
			},
			want: RedisConnectionString{
				Address:  "localhost",
				Password: "password",
				Port:     6379,
				Database: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cli.parseConnectionString(tt.args.connString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedisClient.parseConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockRedis() *miniredis.Miniredis {
	s, err := miniredis.Run()

	if err != nil {
		panic(err)
	}

	return s
}

func setup() *RedisClient {
	redisServer = mockRedis()

	return New(redisServer.Addr())
}

func teardown() {
	redisServer.Close()
}

func TestRedisClient_Set(t *testing.T) {
	client := setup()

	defer teardown()

	err := client.SetString("foo", "bar")

	assert.Nilf(t, err, "we expect the error to be nil")
}

func TestRedisClient_SetExpiring(t *testing.T) {
	client := setup()

	defer teardown()

	setErr := client.SetExpiringString("foo", "bar", 1*time.Second)
	value, getErr := client.GetStringKey("foo")
	expiresIn := redisServer.TTL("foo")

	assert.Nilf(t, setErr, "we expect the error to be nil")
	assert.Nilf(t, getErr, "we expect the get error to be nil")
	assert.Equalf(t, "bar", value, "we expected 'bar' found %s", value)
	assert.Equalf(t, 1*time.Second, expiresIn, "we expect the message ttl to be 1 second")
}

func TestRedisClient_Get(t *testing.T) {
	client := setup()

	defer teardown()

	setErr := client.SetString("foo", "bar")
	result, getErr := client.GetStringKey("foo")

	assert.Nilf(t, setErr, "we expect the set error to be nil")
	assert.Nilf(t, getErr, "we expect the get error to be nil")
	assert.Equalf(t, "bar", result, "expected 'bar' found %s", result)
}

func TestRedisClient_GetEmpty(t *testing.T) {
	client := setup()

	defer teardown()

	result, getErr := client.GetStringKey("foo")

	assert.Errorf(t, getErr, "the key does not exist", "error is not what expected")
	assert.Equalf(t, "", result, "expected '' found %s", result)
}

func TestRedisClient_GetServerError(t *testing.T) {
	client := setup()
	teardown()

	result, getErr := client.GetStringKey("foo")

	assert.Errorf(t, getErr, "the key does not exist", "error is not what expected")
	assert.Equalf(t, "", result, "expected '' found %s", result)
}

func TestRedisClient_Delete(t *testing.T) {
	client := setup()

	defer teardown()

	setErr := client.SetString("foo", "bar")
	result, getErr := client.GetStringKey("foo")
	count, delErr := client.Delete("foo")
	resultDelete, getDeleteErr := client.GetStringKey("foo")

	assert.Nilf(t, setErr, "we expect the set error to be nil")
	assert.Nilf(t, getErr, "we expect the get error to be nil")
	assert.Equalf(t, "bar", result, "expected 'bar' found %s", result)
	assert.Nilf(t, delErr, "we expect the get error to be nil")
	assert.Equalf(t, int64(1), count, "expected count to be 1")
	assert.Equalf(t, "", resultDelete, "expected the key to be empty after delete")
	assert.Errorf(t, getDeleteErr, "we expect to have an error as key no longer exists")
}

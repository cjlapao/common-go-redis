package redisclient

import (
	"reflect"
	"testing"
)

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

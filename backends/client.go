package backends

import (
	"errors"
	"strings"

	"github.com/zyf0330/confd/backends/consul"
	"github.com/zyf0330/confd/backends/dynamodb"
	"github.com/zyf0330/confd/backends/env"
	"github.com/zyf0330/confd/backends/etcd"
	"github.com/zyf0330/confd/backends/etcdv3"
	"github.com/zyf0330/confd/backends/file"
	"github.com/zyf0330/confd/backends/rancher"
	"github.com/zyf0330/confd/backends/redis"
	"github.com/zyf0330/confd/backends/ssm"
	"github.com/zyf0330/confd/backends/vault"
	"github.com/zyf0330/confd/backends/zookeeper"
	"github.com/zyf0330/confd/log"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcd"
	}
	backendNodes := config.BackendNodes

	if config.Backend == "file" {
		log.Info("Backend source(s) set to " + config.YAMLFile)
	} else {
		log.Info("Backend source(s) set to " + strings.Join(backendNodes, ", "))
	}

	switch config.Backend {
	case "consul":
		return consul.New(config.BackendNodes, config.Scheme,
			config.ClientCert, config.ClientKey,
			config.ClientCaKeys)
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "etcdv3":
		return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "zookeeper":
		return zookeeper.NewZookeeperClient(backendNodes)
	case "rancher":
		return rancher.NewRancherClient(backendNodes)
	case "redis":
		return redis.NewRedisClient(backendNodes, config.ClientKey)
	case "env":
		return env.NewEnvClient()
	case "file":
		return file.NewFileClient(config.YAMLFile)
	case "vault":
		vaultConfig := map[string]string{
			"app-id":   config.AppID,
			"user-id":  config.UserID,
			"username": config.Username,
			"password": config.Password,
			"token":    config.AuthToken,
			"cert":     config.ClientCert,
			"key":      config.ClientKey,
			"caCert":   config.ClientCaKeys,
		}
		return vault.New(backendNodes[0], config.AuthType, vaultConfig)
	case "dynamodb":
		table := config.Table
		log.Info("DynamoDB table set to " + table)
		return dynamodb.NewDynamoDBClient(table)
	case "ssm":
		return ssm.New()
	}
	return nil, errors.New("Invalid backend")
}

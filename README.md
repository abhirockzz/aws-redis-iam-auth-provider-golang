# Authenticate Go apps to Redis on AWS using IAM

You can use this package to authenticate your Go apps to Amazon MemoryDB (and Amazon ElastiCache) for Redis using AWS IAM. Example below

> For more info, refer to this blog post.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/abhirockzz/aws-redis-iam-auth-provider-golang/auth"
	"github.com/redis/go-redis/v9"
)

func main() {

    serviceName := "memorydb" // or "elasticache"
    clusterName := "name of cluster"
    username := "iam user name"
    region := "aws region"

	generator, err := auth.New(serviceName, clusterName, username, region)
	if err != nil {
		log.Fatal("failed to initialise token generator", err)
	}

	client := redis.NewClusterClient(
		&redis.ClusterOptions{
			Username: username,
			Addrs:    []string{clusterEndpoint},
			NewClient: func(opt *redis.Options) *redis.Client {

				return redis.NewClient(&redis.Options{
					Addr: opt.Addr,
					CredentialsProvider: func() (username string, password string) {

						token, err := generator.Generate()
						if err != nil {
							log.Fatal("failed to generate auth token", err)
						}

						return opt.Username, token
					},
					TLSConfig: &tls.Config{InsecureSkipVerify: true},
				})
			},
		})

	err = client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal("failed to connect to memorydb -", err)
	}

	fmt.Println("successfully connected to cluster", clusterEndpoint)
}
```



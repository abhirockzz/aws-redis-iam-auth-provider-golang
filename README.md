# Using IAM authentication for Redis on AWS

You can use this package to authenticate your Go apps to Amazon MemoryDB (and Amazon ElastiCache) for Redis using AWS IAM. 

![](https://community.aws/_next/image?url=https%3A%2F%2Fassets.community.aws%2Fa%2F2ZCVX81lcmA658o2P05GmRPjRCU.jpeg%3FimgSize%3D918x370&w=1920&q=75)

Here is an example:

```go
package main

import (
	"context"
	"crypto/tls"
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
	clusterEndpoint := "cluster endpoint" // memorydb or elasticache endpoint

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

For a deep-dive, refer to [this blog post](https://community.aws/content/2ZCKrwaaaTglCCWISSaKv1d7bI3/using-iam-authentication-for-redis-on-aws).

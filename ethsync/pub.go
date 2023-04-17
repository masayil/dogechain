package ethsync

import (
	"context"
	"fmt"
	"strings"

	"log"

	"github.com/go-redis/redis/v8"
)

type PubKv struct {
	Prefix            string
	TempKV            redis.UniversalClient
	topicHeaderKey    string
	topicLogsKey      string
	topicPendingTxKey string
	chWorkLimit       chan struct{}
}

func NewPub(prefix string, tempKVList []string, password string) *PubKv {
	pbc := &PubKv{
		Prefix:            prefix,
		topicHeaderKey:    strings.Join([]string{prefix, "/sub/header"}, ""),
		topicPendingTxKey: strings.Join([]string{prefix, "/sub/pendingTx"}, ""),
		topicLogsKey:      strings.Join([]string{prefix, "/sub/logs"}, ""),
		chWorkLimit:       make(chan struct{}, 1000),
	}
	if len(tempKVList) == 1 {
		pbc.TempKV = redis.NewClient(&redis.Options{
			Addr:     tempKVList[0],
			Password: password,
		})
	} else if len(tempKVList) > 1 {
		pbc.TempKV = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        tempKVList,
			Password:     password,
			PoolSize:     32,
			MinIdleConns: 32,
		})
	}

	if pbc.TempKV != nil {
		sc := pbc.TempKV.Ping(context.Background())
		if sc.Err() != nil {
			log.Fatal(fmt.Errorf("temporary kv storage service ping failed, err is %s", sc.Err().Error()))
		}
	}

	return pbc
}

func (pbc *PubKv) PublishTopicInTemp(ctx context.Context, topicKey string, message interface{}) error {
	pbc.chWorkLimit <- struct{}{}
	go func(message interface{}) error {
		if pbc.TempKV != nil {
			<-pbc.chWorkLimit
			return pbc.TempKV.Publish(ctx, topicKey, message).Err()
		}
		<-pbc.chWorkLimit
		return nil
	}(message)
	return nil
}

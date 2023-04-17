package ethsync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-hclog"
)

// perm store for ankr

type KvSync struct {
	temporaryExp time.Duration
	Prefix       string
	topicKey     string
	TempKV       redis.UniversalClient
	logger       hclog.Logger
	chWorkLimit  chan struct{}
}

func (kvc *KvSync) MergeKey(key string) string {
	return kvc.mergeKey(key)
}

func NewKv(prefix string, tempKVList []string, temporaryExp time.Duration, logger hclog.Logger, password string) *KvSync {
	kvc := &KvSync{
		Prefix:       prefix,
		temporaryExp: temporaryExp,
		chWorkLimit:  make(chan struct{}, 1000),
		topicKey:     strings.Join([]string{prefix, "/ht/latest"}, ""),
		logger:       logger.Named("ankr kv"),
	}

	if len(tempKVList) == 1 {
		kvc.TempKV = redis.NewClient(&redis.Options{
			Addr:     tempKVList[0],
			Password: password,
		})
	} else if len(tempKVList) > 1 {
		kvc.TempKV = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        tempKVList,
			Password:     password,
			PoolSize:     32,
			MinIdleConns: 32,
		})
	}

	if kvc.TempKV != nil {
		sc := kvc.TempKV.Ping(context.Background())
		if sc.Err() != nil {
			log.Fatal(fmt.Errorf("temporary kv storage service ping failed, err is %s", sc.Err().Error()))
		}
	}
	return kvc
}

func (kvc *KvSync) SetInKV(ctx context.Context, key string, value interface{}, logger hclog.Logger) error {
	// if kvc.PermKV != nil {
	// 	return kvc.PermKV.Set(ctx, kvc.mergeKey(key), value, 0).Err()
	// }

	// var wg sync.WaitGroup
	kvc.chWorkLimit <- struct{}{}
	go func(key string, value interface{}) error {
		// wg.Add(1)
		if kvc.TempKV != nil {
			// logger.Info(fmt.Sprintf("ankr store key is %s , ankr store value is %s", kvc.mergeKey(key), value))
			<-kvc.chWorkLimit
			err := kvc.TempKV.Set(ctx, kvc.mergeKey(key), value, kvc.temporaryExp).Err()
			if err != nil {
				kvc.logger.Error(fmt.Sprintf("ankr store error %s", err.Error()))
			}
			return err
		}
		<-kvc.chWorkLimit
		return nil
	}(key, value)
	//wg.Wait()

	return nil
}

func (kvc *KvSync) mergeKey(key string) string {
	return strings.Join([]string{kvc.Prefix, key}, "")
}

func (kvc *KvSync) SetTopicInTemp(ctx context.Context, value interface{}) {
	if kvc.TempKV != nil {
		kvc.TempKV.Set(ctx, kvc.topicKey, value, 0)
	}
}

func (kvc *KvSync) PublishTopicInTemp(ctx context.Context, message interface{}) error {
	if kvc.TempKV != nil {
		return kvc.TempKV.Publish(ctx, kvc.topicKey, message).Err()
	}
	return nil
}

func (kvc *KvSync) GetInTerm(ctx context.Context, key string) (string, error) {
	if kvc.TempKV != nil {
		sc := kvc.TempKV.Get(ctx, kvc.mergeKey(key))
		return sc.Val(), sc.Err()
	}
	return "", nil
}

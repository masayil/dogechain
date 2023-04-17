package ethsync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type DistributedLockKv struct {
	mutex  sync.Mutex
	TempKV redis.UniversalClient
	key    string
}

func NewDistributedLock(tempKVList []string, password string) *DistributedLockKv {
	dlc := &DistributedLockKv{
		mutex: sync.Mutex{},
		key:   "distributed_lock_key",
	}
	if len(tempKVList) == 1 {
		dlc.TempKV = redis.NewClient(&redis.Options{
			Addr:     tempKVList[0],
			Password: password,
		})
	} else if len(tempKVList) > 1 {
		dlc.TempKV = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        tempKVList,
			Password:     password,
			PoolSize:     32,
			MinIdleConns: 32,
		})
	}
	if dlc.TempKV != nil {
		sc := dlc.TempKV.Ping(context.Background())
		if sc.Err() != nil {
			log.Fatal(fmt.Errorf("dlc temporary kv storage service ping failed, err is %s", sc.Err().Error()))
		}
	}
	return dlc
}

func (dlc *DistributedLockKv) Lock() (bool, error) {
	dlc.mutex.Lock()
	defer dlc.mutex.Unlock()
	return dlc.TempKV.SetNX(context.Background(), dlc.key, 1, 5*time.Second).Result()
}

func (dlc *DistributedLockKv) UnLock() (int64, error) {
	nums, err := dlc.TempKV.Del(context.Background(), dlc.key).Result()
	if err != nil {
		return 0, err
	}
	return nums, err
}

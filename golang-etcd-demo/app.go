package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type DistributedMutex interface {
	Run(key string, prefix string, f func()) error
	Close()
}

type EtcdMutex struct {
	Cli       *clientv3.Client
	GrantTime int
}

func (e *EtcdMutex) Run(key string, prefix string, f func()) error {
	resp, err := e.Cli.Grant(context.TODO(), int64(e.GrantTime)) // 50s
	if err != nil {
		return err
	}
	res, err := e.Cli.Put(context.TODO(), key, "host", clientv3.WithLease((resp.ID)))
	if err != nil {
		return err
	}
	defer func() {
		e.Cli.Delete(context.TODO(), key)
	}()
	// 判断当前key是否有执行权限
	resp1, err := e.Cli.Get(context.TODO(), prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	preKey := ""
	for _, v := range resp1.Kvs {
		if v.ModRevision-res.Header.Revision == -1 {
			preKey = string(v.Key)
			break
		}
	}
	if preKey != "" {
		done := make(chan int, 0)
		go func() {
			// 监听前一个比自己小的key
			rch := e.Cli.Watch(context.Background(), preKey)
			for wresp := range rch {
				for _, ev := range wresp.Events {
					// 删除事件
					if string(ev.Kv.Key) == preKey && ev.Type == 1 {
						close(done)
					}
				}
			}
		}()
		<-done
	}
	f()
	return nil
}

func (e *EtcdMutex) Close() {
	e.Cli.Close()
}

func NewEtcdMutex(host string, timeout time.Duration, grantTime int) DistributedMutex {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{host},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &EtcdMutex{
		Cli:       cli,
		GrantTime: grantTime,
	}

}

func main() {
	etcdMutex := NewEtcdMutex("http://127.0.0.1:2379", 2*time.Second, 50)
	defer etcdMutex.Close()
	share := 0
	// single cycle
	for i := 0; i < 10; i++ {
		prefix := "/lock/mylock"
		key := fmt.Sprintf("%s/%s", prefix, uuid.New().String())
		go etcdMutex.Run(key, prefix, func() {
			share++
			fmt.Println(share)
			time.Sleep(3 * time.Second)
		})
	}
	time.Sleep(1 * time.Hour)
}

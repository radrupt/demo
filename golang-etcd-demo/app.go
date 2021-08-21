package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type DistributedMutex struct {
}

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	_, err = cli.Put(ctx, "sample_key", "wangdi")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	_, err = cli.Get(ctx1, "sample_key")
	cancel1()
	if err != nil {
		log.Fatal(err)
	}
	share := 0
	simulationConcurrent := func(i int) {
		key := fmt.Sprintf("/lock/mylock/%d", i)
		resp, err := cli.Grant(context.TODO(), 50)
		if err != nil {
			log.Fatal(err)
		}
		res, err := cli.Put(context.TODO(), key, "bar", clientv3.WithLease((resp.ID)))
		if err != nil {
			log.Fatal(err)
		}
		// 默认应该是升序排列
		resp1, err := cli.Get(context.TODO(), "/lock/mylock", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		preKey := ""
		for _, v := range resp1.Kvs {
			if v.ModRevision-res.Header.Revision == -1 {
				preKey = string(v.Key)
				break
			}
		}
		done := make(chan int, 0)
		if preKey != "" {
			go func() {
				// 监听前一个比自己小的key
				fmt.Println(preKey)
				rch := cli.Watch(context.Background(), preKey)
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
		// 获取锁
		// 启动续期定时任务
		time.Sleep(2 * time.Second)
		share++
		fmt.Println(share)
		cli.Delete(context.Background(), key)
		return
	}
	// single cycle
	for i := 0; i < 10; i++ {
		go simulationConcurrent(i)
	}
	time.Sleep(1 * time.Hour)
}

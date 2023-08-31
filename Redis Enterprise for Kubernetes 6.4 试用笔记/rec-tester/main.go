package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	REDIS_ADDR   = "cluster1:18811"
	REDIS_PASSWD = "123456"

	setFormat  = "wrype:set:%d"
	hashFormat = "wrype:hash:%d"
	randRange  = 1000
	bmTime     = time.Minute
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func bmSet(ctx context.Context, rdb *redis.Client) {
	t := time.NewTimer(bmTime)
	count := 0
	for {
		select {
		case <-t.C:
			log.Printf("Set %d times in %s, rate: %d op/s\n", count, bmTime.String(), count/int(bmTime.Seconds()))
			return
		default:
			err := rdb.Set(ctx, fmt.Sprintf(setFormat, rnd.Intn(randRange)),
				time.Now().Unix(), time.Minute*5).Err()
			if err != nil {
				log.Panic(err)
			}
			count++
		}
	}
}

func bmGet(ctx context.Context, rdb *redis.Client) {
	t := time.NewTimer(bmTime)
	count := 0
	for {
		select {
		case <-t.C:
			log.Printf("Get %d times in %s, rate: %d op/s\n", count, bmTime.String(), count/int(bmTime.Seconds()))
			return
		default:
			err := rdb.Get(ctx, fmt.Sprintf(setFormat, rnd.Intn(randRange))).Err()
			if err != nil {
				log.Panic(err)
			}
			count++
		}
	}
}

var fileds = []string{
	"tm",
	"user",
	"id",
	"val",
}

func bmHSet(ctx context.Context, rdb *redis.Client) {
	t := time.NewTimer(bmTime)
	count := 0
	for {
		select {
		case <-t.C:
			log.Printf("HSet %d times in %s, rate: %d op/s\n", count, bmTime.String(), count/int(bmTime.Seconds()))
			return
		default:
			err := rdb.HSet(ctx, fmt.Sprintf(hashFormat, rnd.Intn(randRange)),
				fileds[rnd.Intn(len(fileds))], time.Now().UnixNano()).Err()
			if err != nil {
				log.Panic(err)
			}
			count++
		}
	}
}

func bmHGet(ctx context.Context, rdb *redis.Client) {
	t := time.NewTimer(bmTime)
	count := 0
	for {
		select {
		case <-t.C:
			log.Printf("HGet %d times in %s, rate: %d op/s\n", count, bmTime.String(), count/int(bmTime.Seconds()))
			return
		default:
			err := rdb.HGet(ctx, fmt.Sprintf(hashFormat, rnd.Intn(randRange)), fileds[rnd.Intn(len(fileds))]).Err()
			switch err {
			case nil, redis.Nil:
				count++
			default:
				log.Panic(err)
			}
		}
	}
}

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if len(addr) == 0 {
		addr = REDIS_ADDR
	}
	passwd, exist := os.LookupEnv("REDIS_PASSWD")
	if !exist {
		passwd = REDIS_PASSWD
	}
	opts := redis.Options{
		Addr:     addr,
		Password: passwd,
	}
	rdb := redis.NewClient(&opts)
	log.Printf("redis client config: %+v\n", opts)
	pong, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Panic(err)
	}
	log.Println(pong)
	defer rdb.Close()

	bmSet(context.TODO(), rdb)
	bmGet(context.TODO(), rdb)
	bmHSet(context.TODO(), rdb)
	bmHGet(context.TODO(), rdb)

	select {}
}

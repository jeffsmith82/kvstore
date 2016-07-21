package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

type Values struct {
	Value string
	Ttl   int64
}

type Key struct {
	Value string
	sync.RWMutex
}

type Kvstore struct {
	data map[string]Values
}

func (db *Kvstore) Get(key string) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := db.data[key]; ok {
		if time.Now().Unix() <= val.Ttl {
			return val.Value, nil
		}
	}
	return "", errors.New("no value")
}

func (db *Kvstore) Set(key string, value string, ttl int64) {
	mutex.Lock()
	defer mutex.Unlock()
	if ttl == 0 {
		db.data[key] = Values{Value: value, Ttl: ttl}
	} else {
		db.data[key] = Values{Value: value, Ttl: time.Now().Unix() + ttl}
	}
}

func NewKvstore() *Kvstore {
	ndb := new(Kvstore)
	ndb.data = make(map[string]Values)
	return ndb
}

func getkey(value string, wg *sync.WaitGroup, db *Kvstore) {
	value, err := db.Get(value)
	if err != nil {
		fmt.Println("No value for key")
	}
	wg.Done()
	fmt.Println(value)
}

func main() {

	var wg sync.WaitGroup
	wg.Add(5)

	db := NewKvstore()
	db.Set("Key1", "value1", 2)
	db.Set("Key2", "value2", 10)
	db.Set("Key3", "value3", 100)
	db.Set("Key4", "value4", 10)

	getkey("Key1", &wg, db)
	time.Sleep(5 * time.Second)
	go getkey("Key1", &wg, db)
	go getkey("Key2", &wg, db)
	db.Set("Key3", "value3", 100)
	go getkey("Key3", &wg, db)
	go getkey("Key4", &wg, db)

	wg.Wait()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Println(mem.Alloc)
	log.Println(mem.TotalAlloc)
	log.Println(mem.HeapAlloc)
	log.Println(mem.HeapSys)

	now := time.Now()

	for i := 0; i < 1000000; i++ {
		db.Set("Key"+strconv.Itoa(i), "value1", 0)
	}
	fmt.Println(now.Sub(time.Now()))
	fmt.Println("Allkeys added")

	runtime.ReadMemStats(&mem)
	log.Println(mem.Alloc)
	log.Println(mem.TotalAlloc)
	log.Println(mem.HeapAlloc)
	log.Println(mem.HeapSys)
}

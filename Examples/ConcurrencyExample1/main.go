package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cache = map[int]Book{}
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	waitGroup := &sync.WaitGroup{}
	//	mtx := &sync.Mutex{}
	mtx := &sync.RWMutex{}
	cacheCh := make(chan Book)
	dbCh := make(chan Book)
	for i := 0; i < 10; i++ {
		id := rnd.Intn(10) + 1
		waitGroup.Add(2)
		go func(id int, w *sync.WaitGroup, mtx *sync.RWMutex, ch chan<- Book) {
			if b, ok := queryCache(id, mtx); ok {
				ch <- b
			}
			w.Done()
		}(id, waitGroup, mtx, cacheCh)

		go func(id int, w *sync.WaitGroup, mtx *sync.RWMutex, ch chan<- Book) {
			if b, ok := queryDatabase(id, mtx); ok {
				ch <- b
			}
			w.Done()
		}(id, waitGroup, mtx, dbCh)
		go func(cacheCh <-chan Book, dbCh <-chan Book) {
			select {
			case b := <-cacheCh:
				fmt.Println("from cache")
				fmt.Println(b)
				<-dbCh
			case b := <-dbCh:
				fmt.Println("from database")
				fmt.Println(b)
			}

		}(cacheCh, dbCh)
		time.Sleep(150 * time.Millisecond)
	}
	waitGroup.Wait()
}

func queryCache(id int, mtx *sync.RWMutex) (Book, bool) {
	mtx.RLock()
	b, ok := cache[id]
	mtx.RUnlock()
	return b, ok
}

func queryDatabase(id int, mtx *sync.RWMutex) (Book, bool) {
	time.Sleep(100 * time.Millisecond)
	for _, b := range books {
		if b.ID == id {
			mtx.Lock()
			cache[id] = b
			mtx.Unlock()
			return b, true
		}
	}
	return Book{}, false
}

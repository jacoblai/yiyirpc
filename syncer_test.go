package yiyirpc

import (
	"testing"
	"time"
	"fmt"
	"log"
	"io/ioutil"
)

func TestListenRPC(t *testing.T) {
	srv := NewRpcServer()
	srv.Register(NewWorker())
	go srv.ListenRPC(4200)
	N := 10
	mapChan := make(chan int, N)
	bts, _:= ioutil.ReadFile(`/Users/lidonghai/Downloads/NavicatPremium.zip`)
	for i := 0; i < N; i++ {
		go func(i int) {
			client := NewRpcClient(120)
			var rep []byte
			nt := time.Now()
			err := client.Call("localhost:4200", "Worker.DoJob", bts, &rep)
			if err != nil {
				t.Error(err)
			} else {
				sub := time.Now().Sub(nt)
				fmt.Println(i, string(rep), sub)
			}
			mapChan <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-mapChan
	}
}

type Worker struct {
	Name string
}

func NewWorker() *Worker {
	return &Worker{"test"}
}

func (w *Worker) DoJob(task []byte, reply *[]byte) error {
	log.Println("Worker: do job", len(task))
	//time.Sleep(time.Second * 3)
	*reply = []byte("OK")
	return nil
}

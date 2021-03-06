package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// channel == synchronization queue == 同步队列？
// -> 【】 ->

func main() {
	ch := make(chan task) //  connection pool , task pool
	go receiver(ch)

	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sentence, err := buf.ReadBytes('\n')
		if err != nil {
			panic(err)
		}

		if len(sentence) > 2 {
			r := bytes.NewReader(
				sentence[:len(sentence)-2])
			go concurrent(r, ch)
		}
	}
}

func receiver(ch chan task) {
	var mutex sync.Mutex
	var tasks []task

	started := make(chan struct{})

	go func(ch chan task) {
		for {
			t := <-ch
			mutex.Lock()
			tasks = append(tasks, t)
			mutex.Unlock()
			started <- struct{}{}
		}
	}(ch)

	i := 0
	for {
		<-started
		mutex.Lock()
		body := <-tasks[i].body
		mutex.Unlock()
		i++
		io.Copy(os.Stdout, body)
		fmt.Println("??")
		fmt.Print("> ")
	}
}

type task struct {
	body chan io.Reader
}

func concurrent(r io.Reader, taskCh chan task) {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", r)

	// task channel 创建任务的时候，就放入channel
	t := task{
		body: make(chan io.Reader),
	}
	taskCh <- t
	res, err := http.DefaultClient.Do(req) // 一个 Do 就是一个任务
	if err != nil {
		panic(err)
	}
	t.body <- res.Body // 结束的时候，通知该任务我结束了
	// close(t.body)
}

package main

import (
	"fmt"
	"time"
)

func play() {
	c := make(chan interface{})
	time.AfterFunc(time.Second*2, func() {
		close(c)
	})
	e := <-c
	fmt.Println(e)
}

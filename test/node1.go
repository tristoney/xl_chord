package main

import (
	"fmt"
	"github.com/tristoney/xl_chord"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	file, _ := os.OpenFile("../log/node1.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	log.SetOutput(file)//设置输出流
	node, err := xl_chord.SpawnNode("127.0.0.1:50001", "")
	if err != nil {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	_ = node.StoreKeyAPI("1", "one")
	key, _ := node.FindKeyAPI("1")
	fmt.Println(key)
	shut := make(chan bool)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.Print()
			case <-shut:
				ticker.Stop()
				return
			}
		}
	}()
	<-c
	shut <- true
	node.GracefulShutdown()
}



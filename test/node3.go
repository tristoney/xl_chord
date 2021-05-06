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
	file, _ := os.OpenFile("../log/node3.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()

	log.SetOutput(file) //设置输出流
	node, err := xl_chord.SpawnNode("127.0.0.1:50003", "127.0.0.1:50001")
	if err != nil {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	shut := make(chan bool)
	_ = node.StoreKeyAPI("3", "three")
	go func() {
		kInt := 4
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.Print()
				k := fmt.Sprintf("%d", kInt)
				t := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05"))
				if kInt < 10 {
					_ = node.StoreKeyAPI(k, t)
				}

			case <-shut:
				ticker.Stop()
				return
			}
			kInt += 1
		}
	}()
	<-c
	node.GracefulShutdown()
}

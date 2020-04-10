package main

import (
	"bufio"
	"os"
	"shila/appep/tun"
)

var dev tun.Device

func main() {

	quit := make(chan string)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		quit <- text
	}()

	dev.Allocate()

	<-quit
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"shila/appep/tun"
	"time"
)

func main() {

	quit := make(chan string)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		quit <- text
	}()

	dev27 := tun.New("tun27")

	dev27.Name = "tun27"
	if err := dev27.Allocate(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("Allocated ", dev27.Name, ".\n")
	}

	time.Sleep(4 * time.Second)

	if err := dev27.Allocate(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("Allocated ", dev27.Name, ".\n")
	}

	time.Sleep(4 * time.Second)

	if err := dev27.Deallocate(); err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Print("Deallocated ", dev27.Name, ".\n")
	}

	time.Sleep(2 * time.Second)

	if err := dev27.Deallocate(); err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Print("Deallocated ", dev27.Name, ".\n")
	}
	<-quit
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"shila/appep/vif"
)

func main() {

	quit := make(chan string)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		quit <- text
	}()

	// sudo ip netns add shila-ingress
	// sudo ip netns exec shila-ingress ip a
	//dev27 := vif.New("tun27", &vif.Namespace{"shila-ingress"}, "10.0.0.1/24")

	dev27 := vif.New("tun27", &vif.Namespace{"shila-ingress"}, "10.0.0.1/24")

	if err := dev27.TurnUp(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("TurnUp.")
	}

	if err := dev27.Setup(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("SetUp.")
	}
	<-quit

	if err := dev27.TurnUp(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("TurnUp.")
	}
	<-quit

	if err := dev27.Teardown(); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Print("Teardown.")
	}
	<-quit

	return
}

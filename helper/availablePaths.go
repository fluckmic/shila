package main

import (
	"flag"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/addr"
)

func main() {

	destAddrString := flag.String("destAddr", "", "Remote SCION Address (e.g. 17-ffaa:1:1)")
	flag.Parse()

	if destAddr, err := addr.IAFromString(*destAddrString); err != nil {
		fmt.Println(err.Error())
		return
	} else {
		if paths, err := appnet.QueryPaths(destAddr); err != nil {
			fmt.Println(err.Error())
			return
		} else if paths == nil {
			// Destination address is in the local IA
			fmt.Printf("Destination %v in local IA - No paths available.\n", destAddr)
			return
		} else {
			fmt.Printf("Paths for destination %v.\n", destAddr)
			for index, path := range paths  {
				fmt.Printf("[%2d] %s\n", index, fmt.Sprintf("%s", path))
			}
		}
	}
}


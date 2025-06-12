package main

import (
	"ChatRoomLittle/Func"
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "192.168.56.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	Manager := Func.CreateClientManager()
	for {
		Conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go Func.PreWork(Conn, Manager)
	}
}

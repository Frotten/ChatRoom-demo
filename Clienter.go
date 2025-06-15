package main

import (
	"ChatRoomLittle/Func"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	Conn, err := net.Dial("tcp", "192.168.56.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()
	go Func.Receive(Conn)
	for {
		Reader := bufio.NewReader(os.Stdin)
		Temp, err := Reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error Reading")
			break
		}
		Temp = strings.TrimSpace(Temp)
		if strings.ToUpper(Temp) == "EXIT" {
			fmt.Println("EXIT Succeed")
			Conn.Write([]byte("对方已退出连接"))
			break
		}
		Conn.Write([]byte(Temp))
	}
}

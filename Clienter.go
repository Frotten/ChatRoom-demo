package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func Receive(conn net.Conn) {
	for {
		Message := make([]byte, 1024)
		n, err := conn.Read(Message)
		if err != nil && err != io.EOF {
			fmt.Println("Error Reading")
			return
		}
		if n != 0 {
			fmt.Println(string(Message[:n]))
		}
	}
}
func main() {
	Conn, err := net.Dial("tcp", "192.168.56.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()
	go Receive(Conn)
	for {
		Reader := bufio.NewReader(os.Stdin)
		Temp, err := Reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error Reading")
			continue
		}
		Temp = strings.TrimSpace(Temp)
		if strings.ToUpper(Temp) == "EXIT" {
			fmt.Println("EXIT Succeed")
			break
		}
		Conn.Write([]byte(Temp))
	}
}

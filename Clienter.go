package main

import (
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

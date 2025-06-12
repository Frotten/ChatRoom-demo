package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type ClientManager struct {
	list map[string]net.Conn
	Lock sync.Mutex
}

func CreateClientManager() *ClientManager {
	return &ClientManager{
		list: make(map[string]net.Conn),
	}
}

func (man *ClientManager) AddClient(IP string, Conn net.Conn) {
	man.Lock.Lock()
	defer man.Lock.Unlock()
	man.list[IP] = Conn
}

func ChatEachOtherAchieve(Conn, TargetConn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		Temp := make([]byte, 1024)
		n, err := Conn.Read(Temp)
		if err != nil && err != io.EOF {
			fmt.Println("读取数据失败：", err)
			return
		}
		if n != 0 {
			if strings.ToUpper(string(Temp[:n])) == "EXIT" {
				TargetConn.Write([]byte("对方停止了聊天"))
				return
			}
			Info := Conn.RemoteAddr().String() + ":" + string(Temp[:n])
			TargetConn.Write([]byte(Info))
		}
	}
}

func PreWork(Conn net.Conn, Manager *ClientManager) {
	IP := Conn.RemoteAddr().String()
	fmt.Println("New client:", IP)
	Manager.AddClient(IP, Conn)
	_, err := Conn.Write([]byte("请按照格式输入聊天对象的IP：（连接:IP）\n"))
	if err != nil {
		fmt.Println("Write error:", err)
		Conn.Close()
		return
	}
	for {
		Username := make([]byte, 1024)
		n, err := Conn.Read(Username)
		if err != nil {
			fmt.Println("Write error:", err)
			Conn.Close()
			return
		}
		if strings.HasPrefix(string(Username[:n]), "连接:") {
			fmt.Println(IP+"即将与", string(Username[7:n]), "进行交流")
			go ChatEachOther(Manager, Conn, string(Username[7:n]))
			break
		}
	}
}
func ChatEachOther(Manager *ClientManager, Conn net.Conn, Target string) {
	defer Conn.Close()
	Manager.Lock.Lock()
	TargetConn, ok := Manager.list[Target]
	Manager.Lock.Unlock()
	if !ok {
		Conn.Write([]byte("目标客户端不在线或不存在。\n"))
		return
	}
	Conn.Write([]byte("与" + Target + "的聊天已开始"))
	TargetConn.Write([]byte("与" + Conn.RemoteAddr().String() + "的聊天已开始\n"))
	var wg sync.WaitGroup
	wg.Add(2)
	go ChatEachOtherAchieve(Conn, TargetConn, &wg)
	go ChatEachOtherAchieve(TargetConn, Conn, &wg)
	wg.Wait()
}
func main() {
	listener, err := net.Listen("tcp", "192.168.56.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	Manager := CreateClientManager()
	for {
		Conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go PreWork(Conn, Manager)
	}
}

package Func

import (
	"fmt"
	"io"
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

func (man *ClientManager) RemoveClient(IP string) {
	man.Lock.Lock()
	defer man.Lock.Unlock()
	man.list[IP] = nil
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
			ch1 := make(chan bool)
			go ChatEachOther(Manager, Conn, string(Username[7:n]), ch1)
			<-ch1
		}
	}
}

func ChatEachOther(Manager *ClientManager, Conn net.Conn, Target string, ch1 chan bool) {
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
	go ChatEachOtherAchieve(Conn, TargetConn, &wg, Manager)
	go ChatEachOtherAchieve(TargetConn, Conn, &wg, Manager)
	wg.Wait()
	ch1 <- true
}

func ChatEachOtherAchieve(Conn, TargetConn net.Conn, wg *sync.WaitGroup, Manager *ClientManager) {
	defer wg.Done()
	for {
		Temp := make([]byte, 1024)
		n, err := Conn.Read(Temp)
		if err != nil && err != io.EOF {
			fmt.Println("读取数据失败：", err)
			return
		}
		if n != 0 {
			if strings.HasSuffix(string(Temp[:n]), "对方已退出连接") {
				TargetConn.Write([]byte("对方已退出，聊天结束"))
				Manager.RemoveClient(Conn.RemoteAddr().String())
				break
			}
			Info := Conn.RemoteAddr().String() + ":" + string(Temp[:n])
			TargetConn.Write([]byte(Info))
		}
	}
}

func Receive(conn net.Conn) {
	for {
		Message := make([]byte, 1024)
		n, err := conn.Read(Message)
		if err != nil && err != io.EOF {
			return
		}
		if n != 0 {
			fmt.Println(string(Message[:n]))
		}
	}
}

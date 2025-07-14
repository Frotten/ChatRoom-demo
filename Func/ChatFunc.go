package Func

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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
	for {
		if Manager.list[IP] == nil {
			break
		}
		_, err := Conn.Write([]byte("请按照格式输入聊天对象的IP：（连接:IP）\n"))
		if err != nil {
			fmt.Println("Write error:", err)
			Conn.Close()
			return
		}
		Username := make([]byte, 1024)
		n, err := Conn.Read(Username)
		if err != nil {
			fmt.Println("Write error:", err)
			Conn.Close()
			return
		}
		if strings.HasPrefix(string(Username[:n]), "连接:") {
			fmt.Println(IP+"即将与", string(Username[7:n]), "进行交流")
			ChatEachOther(Manager, Conn, string(Username[7:n]))
		}
	}
}

func ChatEachOther(Manager *ClientManager, Conn net.Conn, Target string) {
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
	go ChatEachOtherAchieve(Conn.RemoteAddr().String(), Target, &wg, Manager)
	go ChatEachOtherAchieve(Target, Conn.RemoteAddr().String(), &wg, Manager)
	wg.Wait()
	fmt.Println("Over")
}

func ChatEachOtherAchieve(IP1, IP2 string, wg *sync.WaitGroup, Manager *ClientManager) {
	defer wg.Done()
	for {
		Conn := Manager.list[IP1]
		TargetConn := Manager.list[IP2]
		if Conn == nil || TargetConn == nil {
			fmt.Println("其中一方退出，聊天结束")
			return
		}
		Temp := make([]byte, 1024)
		n, err := Conn.Read(Temp)
		if err != nil && err != io.EOF {
			fmt.Println("读取数据失败：", err)
			return
		} else if err == io.EOF {
			fmt.Println("无输入")
			return
		}
		if n != 0 {
			if strings.HasSuffix(string(Temp[:n]), "对方已退出连接") {
				TargetConn.Write([]byte("对方已退出连接"))
				Manager.RemoveClient(Conn.RemoteAddr().String())
				Conn.Close()
				Conn = nil
				return
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

func CreateTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS Client (
	    id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(12) NOT NULL UNIQUE,
		password VARCHAR(12) NOT NULL UNIQUE
	);`
	_, err := db.Exec(query)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return err
	}
	fmt.Println("Table created successfully")
	return nil
}

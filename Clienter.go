package main

import (
	"ChatRoomLittle/Func"
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	Conn, err := net.Dial("tcp", "192.168.56.1:8080") //192.168.1.105
	Starttime := time.Now()
	th := make(chan time.Time)
	end := make(chan bool)
	defer close(end)
	if err != nil {
		log.Fatal(err)
	}
	defer Conn.Close()
	ch1 := make(chan int)
	go Func.TimeOut(Conn, th, end, Starttime)
	go Func.Receive(Conn, ch1)
	for {
		if Conn == nil || th == nil {
			fmt.Println("连接已断开")
			break
		}
		Reader := bufio.NewReader(os.Stdin)
		Temp, err := Reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error Reading")
			break
		}
		if Conn == nil || th == nil {
			fmt.Println("连接已断开")
			break
		}
		Temp = strings.TrimSpace(Temp)
		if strings.ToUpper(Temp) == "EXIT" {
			fmt.Println("EXIT Succeed")
			Conn.Write([]byte("\\\\exit"))
			break
		} else if Temp == "\\4" {
			Conn.Write([]byte(Temp))
			Reader01 := bufio.NewReader(os.Stdin)
			fmt.Println("校验中···")
			a := <-ch1
			if a == 1 {
				fmt.Println("校验失败.")
				continue
			} else if a == 2 {
				FilePath, _ := Reader01.ReadString('\n')
				FilePath = strings.TrimSpace(FilePath)
				file01, err := os.Open(FilePath)
				if err != nil {
					fmt.Println("打开文件失败:", err)
					_, _ = Conn.Write([]byte("打开失败"))
					continue
				}
				fileinfo, err := os.Stat(FilePath)
				if err != nil {
					fmt.Println("打开文件失败:", err)
					_, _ = Conn.Write([]byte("打开失败"))
					continue
				}
				size := fileinfo.Size()
				fmt.Println("请输入文件名及后缀:")
				FileName, _ := Reader01.ReadString('\n')
				FileName = strings.TrimSpace(FileName)
				Ans := strconv.Itoa(int(size)) + "|" + FileName
				Conn.Write([]byte(Ans))
				Temp := make([]byte, 1024)
				for {
					n, err := file01.Read(Temp)
					if err != nil {
						if err == io.EOF {
							fmt.Println("文件读取完毕")
							break
						} else {
							fmt.Println("文件读取错误:", err)
							break
						}
					}
					if n > 0 {
						Conn.Write(Temp[:n])
					}
				}
			}
		} else if Temp == "\\5" {
			Conn.Write([]byte(Temp))
			Reader01 := bufio.NewReader(os.Stdin)
			fmt.Println("校验中···")
			a := <-ch1
			if a == 1 {
				fmt.Println("校验失败.")
				continue
			} else if a == 2 {
				fmt.Println("出现异常")
				continue
			} else if a == 3 {
				err := os.MkdirAll("D:\\TryDir\\TempDownload", 0777)
				if err != nil && !os.IsExist(err) {
					fmt.Println("文件夹创建失败，请检查D盘权限")
				}
				Ans, err := Reader01.ReadString('\n')
				if err != nil {
					fmt.Println("Error Reading")
					continue
				}
				Ans = strings.TrimSpace(Ans)
				Conn.Write([]byte(Ans)) //接收数据，然后存放在指定位置
			}
		} else {
			Conn.Write([]byte(Temp))
		}
		select {
		case th <- time.Now():
		case <-end:
			fmt.Println("连接已断开")
			return
		}
	}
}

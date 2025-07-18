package Func

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ClientManager struct {
	list        map[string]net.Conn
	Lock        sync.Mutex
	OnlineUsers int
}

func CreateClientManager() *ClientManager {
	return &ClientManager{
		list: make(map[string]net.Conn),
	}
}

func (man *ClientManager) AddClient(Conn net.Conn, UserName string) {
	man.Lock.Lock()
	defer man.Lock.Unlock()
	man.list[UserName] = Conn
	man.OnlineUsers++
	for _, Conn := range man.list {
		if Conn != nil {
			_, _ = Conn.Write([]byte("用户" + UserName + "已上线,当前在线人数：" + strconv.Itoa(man.OnlineUsers) + "\n"))
		}
	}
}

func (man *ClientManager) RemoveClient(IP string) {
	man.Lock.Lock()
	defer man.Lock.Unlock()
	man.list[IP] = nil
	man.OnlineUsers--
	for _, Conn := range man.list {
		if Conn != nil {
			_, _ = Conn.Write([]byte("用户" + IP + "已下线,当前在线人数：" + strconv.Itoa(man.OnlineUsers) + "\n"))
		}
	}
}

func PreWork(db *sql.DB, Conn net.Conn, Manager *ClientManager) {
	IP := Conn.RemoteAddr().String()
	fmt.Println("New client:", IP)
	tag := true
	Temp, err := os.ReadFile("D:\\Code\\GoCode\\ChatRoomLittle\\提示语.txt")
	if err != nil {
		fmt.Println("Write error:", err)
		_ = Conn.Close()
		return
	}
	for {
		if Conn == nil {
			return
		}
		if tag == true {
			_, err = Conn.Write(Temp)
			if err != nil {
				fmt.Println("Write error:", err)
				_ = Conn.Close()
				return
			}
			tag = false
		}
		Command := make([]byte, 1024)
		n, err := Conn.Read(Command)
		if err != nil {
			fmt.Println("Write error:", err)
			_ = Conn.Close()
			return
		}
		if strings.HasPrefix(string(Command[:n]), "\\") {
			TempString := string(Command[:n])
			switch TempString {
			case "\\1":
				_, _ = Conn.Write([]byte("请输入用户名和密码:\n"))
				_, _ = Conn.Write([]byte("用户名："))
				TempString02 := make([]byte, 1024)
				n1, _ := Conn.Read(TempString02)
				if n1 == 0 {
					_, _ = Conn.Write([]byte("用户名输入错误\n"))
					tag = true
					continue
				} else {
					UserName := strings.TrimSpace(string(TempString02[:n1]))
					_, _ = Conn.Write([]byte("密码："))
					TempString03 := make([]byte, 1024)
					n2, _ := Conn.Read(TempString03)
					if n2 == 0 {
						fmt.Println("密码输入为空,输入无效")
						tag = true
						continue
					} else {
						Password := strings.TrimSpace(string(TempString03[:n2]))
						err = insertUser(db, UserName, Password)
						if err != nil {
							_, _ = Conn.Write([]byte("注册新用户失败，可能是该用户名已存在\n"))
							continue
						}
						_, _ = Conn.Write([]byte("注册成功！\n"))
						tag = true
					}
				}
			case "\\2":
				_, _ = Conn.Write([]byte("请输入用户名和密码:\n"))
				username := make([]byte, 1024)
				password := make([]byte, 1024)
				_, _ = Conn.Write([]byte("用户名："))
				n1, _ := Conn.Read(username)
				_, _ = Conn.Write([]byte("密码："))
				n2, _ := Conn.Read(password)
				Username := strings.TrimSpace(string(username[:n1]))
				Password := strings.TrimSpace(string(password[:n2]))
				Password, _ = Md5(Password)
				rows, err := QueryUser(db)
				if err != nil {
					fmt.Println("查询已有用户失败，疑似发生未知错误")
					_, _ = Conn.Write([]byte("查询已有用户失败，疑似发生未知错误\n"))
					_ = Conn.Close()
					return
				}
				flag := false
				for rows.Next() {
					var stringsA, stringsB string
					_ = rows.Scan(&stringsA, &stringsB)
					if strings.TrimSpace(stringsA) == Username && strings.TrimSpace(stringsB) == Password {
						if Manager.list[Username] != nil {
							_, _ = Conn.Write([]byte("账号已在线，请勿重复登录\n"))
							tag = true
							flag = true
							break
						}
						_, _ = Conn.Write([]byte("账号验证成功，欢迎上线\n"))
						Manager.AddClient(Conn, stringsA)
						AfterLogin(Conn, db, Manager, stringsA)
						tag = true
						flag = true
					}
				}
				if flag == false {
					_, _ = Conn.Write([]byte("账号验证失败，请重新输入\n"))
				}
				_ = rows.Close()
			case "\\4":
			case "\\5":
				_, _ = Conn.Write([]byte("错误：尚未登陆\\\\"))
			case "\\kodayo":
				tag = true
				continue
			default:
				_, _ = Conn.Write([]byte("请先登录"))
			}
		}
	}
}

func AfterLogin(Conn net.Conn, db *sql.DB, Manager *ClientManager, PID string) {
	Temp, _ := os.ReadFile("D:\\Code\\GoCode\\ChatRoomLittle\\提示语2.txt")
	_, _ = Conn.Write(Temp)
	tag := false
	ID := PID
	for {
		if Conn == nil || Manager.list[ID] == nil {
			Manager.RemoveClient(ID)
			return
		}
		if tag == true {
			_, _ = Conn.Write(Temp)
			tag = false
		}
		Words := make([]byte, 1024)
		n, err := Conn.Read(Words)
		if err != nil && err != io.EOF {
			fmt.Println("读取数据失败：", err, "请重新登录")
			Manager.RemoveClient(ID)
			_ = Conn.Close()
			return
		}
		if n != 0 {
			TempString := string(Words[:n])
			switch TempString {
			case "\\1":
				_, _ = Conn.Write([]byte("请输入当前密码："))
				NowPassword := make([]byte, 1024)
				n1, _ := Conn.Read(NowPassword)
				Now, _ := Md5(string(NowPassword[:n1]))
				rows, err := QueryUser(db)
				if err != nil {
					fmt.Println("查询已有用户失败，疑似发生未知错误")
					_, _ = Conn.Write([]byte("查询已有用户失败，疑似发生未知错误\n"))
					Manager.RemoveClient(ID)
					_ = Conn.Close()
					return
				}
				for rows.Next() {
					var stringsA, stringsB string
					_ = rows.Scan(&stringsA, &stringsB)
					if stringsA == ID && stringsB == Now {
						_, _ = Conn.Write([]byte("请输入新密码："))
						NewPassword := make([]byte, 1024)
						n2, _ := Conn.Read(NewPassword)
						if n2 == 0 {
							_, _ = Conn.Write([]byte("新密码输入错误，停止修改\n"))
							tag = true
							break
						} else {
							query := "UPDATE Client SET password = ? WHERE username = ?"
							New, _ := Md5(string(NewPassword[:n2]))
							_, err := db.Exec(query, New, ID)
							if err != nil {
								_, _ = Conn.Write([]byte("更新密码失败，疑似发生未知错误\n"))
								Manager.RemoveClient(ID)
								_ = Conn.Close()
								return
							}
							_, _ = Conn.Write([]byte("密码修改成功！,即将重新进入登陆界面\n"))
							Manager.RemoveClient(ID)
							return
						}
					}
				}
				_, _ = Conn.Write([]byte("当前密码输入错误，停止修改\n"))
				tag = true
				_ = rows.Close()
			case "\\2":
				_, _ = Conn.Write([]byte("请输入新的用户名："))
				NewUserName := make([]byte, 1024)
				n1, _ := Conn.Read(NewUserName)
				rows, err := QueryUser(db)
				if err != nil {
					fmt.Println("查询已有用户失败，疑似发生未知错误")
					_, _ = Conn.Write([]byte("查询已有用户失败，疑似发生未知错误\n"))
					Manager.RemoveClient(ID)
					_ = Conn.Close()
					return
				}
				for rows.Next() {
					var stringsA, stringsB string
					_ = rows.Scan(&stringsA, &stringsB)
					if stringsA == string(NewUserName[:n1]) {
						_, _ = Conn.Write([]byte("用户名已存在，请重新输入\n"))
						tag = true
						break
					}
				}
				if tag == false {
					query := "UPDATE Client SET username = ? WHERE username = ?"
					_, _ = db.Exec(query, string(NewUserName[:n1]), ID)
					Manager.Lock.Lock()
					TempConn := Manager.list[ID]
					Manager.list[ID] = nil
					delete(Manager.list, ID)
					Manager.list[string(NewUserName[:n1])] = TempConn
					Manager.Lock.Unlock()
					RenameQuery := fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", ID, string(NewUserName[:n1]))
					_, _ = db.Exec(RenameQuery)
					_, _ = Conn.Write([]byte("用户名修改完成\n"))
					ID = string(NewUserName[:n1])
				}
				_ = rows.Close()
			case "\\3":
				_, _ = Conn.Write([]byte("当前在线用户列表："))
				Manager.Lock.Lock()
				for ID := range Manager.list {
					if Manager.list[ID] != nil {
						_, _ = Conn.Write([]byte(ID + "\n"))
					}
				}
				Manager.Lock.Unlock()
			case "\\4":
				err := os.MkdirAll("D:\\TryDir\\TempCopy", 0777)
				if err != nil && !os.IsExist(err) {
					_, _ = Conn.Write([]byte("创建目录失败，请检查权限或路径是否正确\n"))
					continue
				}
				_, _ = Conn.Write([]byte("请输入文件绝对路径:\\\\"))
				TempInfo := make([]byte, 1024)
				n1, _ := Conn.Read(TempInfo)
				Finger := strings.Index(string(TempInfo[:n1]), "|")
				Size, _ := strconv.Atoi(string(TempInfo[:Finger]))
				Filename := string(TempInfo[Finger+1 : n1])
				file01, _ := os.OpenFile("D:\\TryDir\\TempCopy\\"+Filename, os.O_CREATE|os.O_RDWR, os.ModePerm)
				Temp := make([]byte, 1024)
				fmt.Println("size:", Size)
				for {
					n2, err := Conn.Read(Temp)
					Size -= n2
					if err != nil || n2 == 0 || Size <= 0 {
						if err == io.EOF || n2 == 0 || Size <= 0 {
							fmt.Println("文件传输结束")
							_ = file01.Close()
							for _, TargetConn := range Manager.list {
								if TargetConn != nil {
									_, _ = TargetConn.Write([]byte("[" + ID + "] 上传了名为：" + Filename + "的文件"))
								}
							}
							break
						} else {
							fmt.Println("读取文件失败:", err)
							_ = file01.Close()
							_, _ = Conn.Write([]byte("读取文件失败\n"))
							break
						}
					}
					if n2 > 0 {
						_, _ = file01.Write(Temp[:n2])
					}
				}
			case "\\5":
				BasicLocation := "D:\\TryDir\\TempCopy"
				DirEntry, err := os.ReadDir(BasicLocation)
				if err != nil {
					_, _ = Conn.Write([]byte("读取目录失败，请检查权限或路径是否正确\n"))
					continue
				}
				mp1 := make(map[int]string)
				for id, entry := range DirEntry {
					_, _ = Conn.Write([]byte(strconv.Itoa(id) + " : " + entry.Name() + "\n"))
					mp1[id] = entry.Name()
				}
				if len(mp1) == 0 {
					_, err = Conn.Write([]byte("暂无可下载文件"))
					if err != nil {
						fmt.Println("发送数据失败:", err)
						_ = Conn.Close()
						return
					}
					_, _ = Conn.Write([]byte("暂无文件在BasicLocation中，请等待添加文件后在进行下载\\\\"))
					continue
				}
				_, _ = Conn.Write([]byte("请输入要下载的文件编号：\\\\"))
				Temp := make([]byte, 1024)
				n1, _ := Conn.Read(Temp)
				Index, err := strconv.Atoi(string(Temp[:n1]))
				if mp1[Index] == "" || err != nil {
					Conn.Write([]byte("错误标签，请确保输入的编号正确"))
					continue
				}
				TargetName := BasicLocation + "\\" + mp1[Index] //这里要把数据传输到客户端上
				file01, err := os.Open(TargetName)
				if err != nil {
					fmt.Println("打开文件失败:", err)
					continue
				}
				fileinfo, _ := file01.Stat()
				_, _ = Conn.Write([]byte("[[即将开始传输文件]]"))
				_, _ = Conn.Write([]byte(strconv.Itoa(int(fileinfo.Size())) + "|" + fileinfo.Name()))
				Temp02 := make([]byte, 1024)
				for {
					n2, err := file01.Read(Temp02)
					if err != nil {
						if err == io.EOF || n2 == 0 {
							_ = file01.Close()
							break
						} else {
							fmt.Println("读取文件失败:", err)
							_ = file01.Close()
							_, _ = Conn.Write([]byte("读取文件失败\n"))
							break
						}
					}
					if n2 > 0 {
						_, _ = Conn.Write(Temp02[:n2])
					}
				}
			case "\\7":
				_, _ = Conn.Write([]byte("当前黑名单列表："))
				query := fmt.Sprintf("SELECT Personnel FROM `%s`", ID)
				Manager.Lock.Lock()
				rows, _ := db.Query(query)
				for rows.Next() {
					var Personnel int
					var PersonnelName string
					_ = rows.Scan(&Personnel)
					idquery := `SELECT username FROM Client WHERE id = ?`
					_ = db.QueryRow(idquery, Personnel).Scan(&PersonnelName)
					_, _ = Conn.Write([]byte(PersonnelName))
				}
				Manager.Lock.Unlock()
			case "\\kodayo":
				tag = true
				continue
			case "\\\\exit":
				Manager.RemoveClient(ID)
				_ = Conn.Close()
				return
			default:
				if strings.HasPrefix(TempString, "@") {
					Index := strings.Index(TempString, "|")
					if Index == -1 {
						_, _ = Conn.Write([]byte("私信格式错误，请确保指令格式正确"))
						continue
					}
					TargetID := strings.TrimSpace(TempString[1:Index])
					TargetConn := Manager.list[TargetID]
					if TargetConn == nil {
						_, _ = Conn.Write([]byte("用户" + TargetID + "不在线或不存在\n"))
					} else {
						query := fmt.Sprintf("SELECT Personnel FROM `%s`", TargetID)
						rows, _ := db.Query(query)
						AllowSympol := true
						for rows.Next() {
							var Personnel int
							_ = rows.Scan(&Personnel)
							IDquery := `SELECT username FROM Client WHERE id = ?`
							var BlackName string
							_ = db.QueryRow(IDquery, Personnel).Scan(&BlackName)
							if BlackName == ID {
								AllowSympol = false
								break
							}
						}
						if AllowSympol == true {
							_, _ = TargetConn.Write([]byte("私信： [" + ID + "] : " + TempString[Index+1:]))
						}
					}
				} else if strings.HasPrefix(TempString, "\\6") {
					Index := strings.Index(TempString, " ")
					if Index == -1 {
						_, _ = Conn.Write([]byte("黑名单添加格式错误，请确保指令与用户名之间留有空格"))
						continue
					}
					TargetID := strings.TrimSpace(TempString[Index+1:])
					Manager.Lock.Lock()
					query := `SELECT id,username FROM Client WHERE username = ?`
					rows, err := db.Query(query, TargetID)
					if err != nil {
						_, _ = Conn.Write([]byte("查询用户失败，疑似发生未知错误"))
						fmt.Println("数据库访问失败")
						Manager.Lock.Unlock()
						_ = rows.Close()
						_ = Conn.Close()
						return
					}
					if rows.Next() {
						var username string
						var id int
						_ = rows.Scan(&id, &username)
						if username == ID {
							_, _ = Conn.Write([]byte("不能将自己加入黑名单\n"))
							continue
						} else {
							blacklist := fmt.Sprintf("INSERT INTO `%s` (Personnel) VALUES (?)", ID)
							_, _ = db.Exec(blacklist, id)
							_, _ = Conn.Write([]byte("用户 " + TargetID + " 已加入黑名单"))
						}
					}
					_ = rows.Close()
					Manager.Lock.Unlock()
				} else if strings.HasPrefix(TempString, "\\8") {
					Index := strings.Index(TempString, " ")
					if Index == -1 {
						_, _ = Conn.Write([]byte("黑名单删除格式错误，请确保指令与用户名之间留有空格"))
						continue
					}
					TargetID := strings.TrimSpace(TempString[Index+1:])
					var TargetNum int
					Manager.Lock.Lock()
					IDquery := `SELECT id FROM client WHERE username = ?`
					_ = db.QueryRow(IDquery, TargetID).Scan(&TargetNum)
					query := fmt.Sprintf("DELETE FROM `%s` WHERE Personnel = ?", ID)
					result, _ := db.Exec(query, TargetNum)
					Manager.Lock.Unlock()
					rowAffect, _ := result.RowsAffected()
					if rowAffect > 0 {
						_, _ = Conn.Write([]byte("用户 " + TargetID + " 已从黑名单中删除"))
					} else {
						_, _ = Conn.Write([]byte("用户 " + TargetID + " 不在黑名单中，删除失败"))
					}
				} else {
					for name, TargetConn := range Manager.list {
						if TargetConn != nil && TargetConn != Conn {
							query := fmt.Sprintf("SELECT Personnel FROM `%s`", name)
							rows, _ := db.Query(query)
							AllowSympol := true
							for rows.Next() {
								var Personnel int
								_ = rows.Scan(&Personnel)
								IDquery := `SELECT username FROM Client WHERE id = ?`
								var BlackName string
								_ = db.QueryRow(IDquery, Personnel).Scan(&BlackName)
								if BlackName == ID {
									AllowSympol = false
									break
								}
							}
							if AllowSympol == true {
								_, _ = TargetConn.Write([]byte("[" + ID + "] : " + TempString))
							}
						}
					}
				}
			}
		}
	}
}

func Receive(conn net.Conn, ch1 chan int) {
	for {
		Message := make([]byte, 1024)
		n, err := conn.Read(Message)
		if err != nil && err != io.EOF {
			fmt.Println("读取数据失败：", err)
			return
		}
		if n != 0 {
			if string(Message[:n]) == "错误：尚未登陆\\\\" {
				ch1 <- 1
				fmt.Println("请先登录")
			} else if string(Message[:n]) == "请输入文件绝对路径:\\\\" {
				ch1 <- 2
				fmt.Println("请输入文件路径：")
			} else if string(Message[:n]) == "请输入要下载的文件编号：\\\\" {
				ch1 <- 3
				fmt.Println("请输入要下载的文件编号：")
			} else if string(Message[:n]) == "暂无文件在BasicLocation中，请等待添加文件后在进行下载\\\\" {
				ch1 <- 4
			} else if string(Message[:n]) == "[[即将开始传输文件]]" {
				n, _ = conn.Read(Message)
				Ans := string(Message[:n])
				Finger := strings.Index(Ans, "|")
				Size, _ := strconv.Atoi(Ans[:Finger])
				Title := Ans[Finger+1 : n]
				file01, _ := os.OpenFile("D:\\TryDir\\TempDownload\\"+Title, os.O_CREATE|os.O_RDWR, os.ModePerm)
				Temp := make([]byte, 1024)
				for {
					n1, err := conn.Read(Temp)
					Size -= n1
					if err != nil || n1 == 0 || Size <= 0 {
						if err == io.EOF || n1 == 0 || Size <= 0 {
							fmt.Println("文件传输结束")
							_ = file01.Close()
							break
						}
					}
					if n1 > 0 {
						_, _ = file01.Write(Temp[:n1])
					}
				}
			} else {
				fmt.Println(string(Message[:n]))
			}
		}
	}
}

func TimeOut(Conn net.Conn, th1 chan time.Time, end chan bool, StartTime time.Time) {
	NowTime := StartTime
	for {
		select {
		case NewTime := <-th1:
			NowTime = NewTime
		default:
			if time.Since(NowTime).Seconds() > 300 {
				fmt.Println("待机超时，已自动退出")
				_ = Conn.Close()
				end <- true
				close(th1)
				return
			}
		}
	}
}
func CreateTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS Client (
	    id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(50) NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return err
	}
	fmt.Println("Table created successfully")
	return nil
}

func insertUser(db *sql.DB, username, password string) error {
	query := "INSERT INTO Client (username, password) VALUES (?, ?)"
	password, _ = Md5(password)
	fmt.Println("Password:", password)
	_, err := db.Exec(query, username, password)
	if err != nil {
		fmt.Println("Error inserting user:", err)
		return err
	}
	create := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (id INT AUTO_INCREMENT PRIMARY KEY, Personnel INT NOT NULL)", username)
	_, err = db.Exec(create)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return err
	}
	fmt.Println("User inserted successfully")
	return nil
}

func QueryUser(db *sql.DB) (*sql.Rows, error) {
	query := "SELECT username, password FROM Client"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func Md5(message string) (string, error) {
	m := md5.New()
	_, err := io.WriteString(m, message)
	if err != nil {
		return "", err
	}
	arr := m.Sum(nil)
	return fmt.Sprintf("%x", arr), nil
}

package main

import (
	"ChatRoomLittle/Func"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	Manager := Func.CreateClientManager()
	db, err := sql.Open("mysql", "Cheter:1234@tcp(192.168.56.1:3306)/ChatRoom")
	if err != nil {
		_ = db.Close()
		log.Fatal("Database connection error:", err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	err = Func.CreateTable(db)
	for {
		Conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go Func.PreWork(db, Conn, Manager)
	}
}

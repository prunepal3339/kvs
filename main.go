package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/prunepal3339/kvs/handler"
	"github.com/prunepal3339/kvs/persistence"
	"github.com/prunepal3339/kvs/resp"
)

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on port :6379")

	aof, err := persistence.NewAof("kvsdb.aof")
	if err != nil {
		fmt.Println(err)
		return
	}

	aof.Read(func(value resp.Value) {
		values := value.Val().([]resp.Value)
		command := strings.ToUpper(values[0].Val().(string))
		args := values[1:]

		handlerFunc, ok := handler.Handlers[command]

		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handlerFunc(args)
	})
	defer aof.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(conn, aof)
	}
}
func handleConnection(conn net.Conn, aof *persistence.Aof) {
	defer conn.Close()
	for {
		reader := resp.NewResp(conn)
		value, err := reader.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		if value.Tag() != resp.TAG_ARR {
			fmt.Println("Invalid request, expected array as root value.")
			continue
		}
		arrayValue := value.Val().([]resp.Value)
		if len(arrayValue) == 0 {
			fmt.Println("Invalid request, expected array to be non empty.")
			continue
		}
		command := strings.ToUpper(arrayValue[0].Val().(string))

		args := arrayValue[1:] // []resp.Value
		handler, ok := handler.Handlers[command]

		writer := resp.NewWriter(conn)

		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(resp.NewValue(resp.TAG_STR, ""))
			continue
		}

		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

		result := handler(args)
		writer.Write(result)
	}
}

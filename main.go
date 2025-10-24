package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/prunepal3339/kvs/handler"
	"github.com/prunepal3339/kvs/resp"
)

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on port :6379")
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
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
		result := handler(args)
		writer.Write(result)
	}
}

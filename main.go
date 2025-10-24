package main

import (
	"fmt"
	"net"

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
		fmt.Println(value)

		writer := resp.NewWriter(conn)
		writer.Write(resp.NewValue(resp.TAG_STR, "Ok"))
	}
}

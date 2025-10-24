package handler

import (
	"github.com/prunepal3339/kvs/resp"
)

var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":    ping,
	"COMMAND": command,
	"GET":     get,
	"SET":     set,
	"HGET":    hget,
	"HSET":    hset,
	"HGETALL": hgetall,
}

var mySet = make(map[string]any)
var myHSet = make(map[string]map[string]any)

func ping(args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.NewValue(resp.TAG_STR, "PONG")
	}
	return resp.NewValue(resp.TAG_STR, args[0].Val())
}
func command(args []resp.Value) resp.Value {
	return resp.NewValue(resp.TAG_STR, "")
}

func get(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 1 argument for GET command")
	}
	key := args[0].Val().(string)
	value, ok := mySet[key]
	if !ok {
		return resp.NewValue(resp.TAG_NIL, nil)
	}
	return resp.NewValue(resp.TAG_BULK, value)
}
func set(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 2 arguments for SET command")
	}
	key := args[0].Val().(string)
	value := args[1].Val()
	mySet[key] = value
	return resp.NewValue(resp.TAG_STR, "OK")
}

func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 2 arguments for HGET command")
	}
	hash := args[0].Val().(string)
	key := args[1].Val().(string)
	value, ok := myHSet[hash][key]
	if !ok {
		return resp.NewValue(resp.TAG_NIL, nil)
	}
	return resp.NewValue(resp.TAG_BULK, value)
}
func hset(args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 3 arguments for HSET command")
	}
	hash := args[0].Val().(string)
	key := args[1].Val().(string)
	value := args[2].Val()
	if myHSet[hash] == nil {
		myHSet[hash] = make(map[string]any)
	}
	myHSet[hash][key] = value

	return resp.NewValue(resp.TAG_STR, "OK")
}
func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 1 argument for HGETALL command")
	}
	hash := args[0].Val().(string)
	value := myHSet[hash]
	var arrayValue []resp.Value
	for k, v := range value {
		arrayValue = append(arrayValue, resp.NewValue(resp.TAG_STR, k))
		arrayValue = append(arrayValue, resp.NewValue(resp.TAG_STR, v))
	}
	return resp.NewValue(resp.TAG_ARR, arrayValue)
}

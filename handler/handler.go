package handler

import (
	"sync"

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

type SafeMap struct {
	sync.RWMutex
	data map[string]any
}
type SafeHMap struct {
	sync.RWMutex
	data map[string]map[string]any
}

var mySet = &SafeMap{data: make(map[string]any)}
var myHSet = &SafeHMap{data: make(map[string]map[string]any)}

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
	mySet.RLock()
	value, ok := mySet.data[key]
	mySet.RUnlock()
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
	mySet.Lock()
	mySet.data[key] = value
	mySet.Unlock()
	return resp.NewValue(resp.TAG_STR, "OK")
}

func hget(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 2 arguments for HGET command")
	}
	hash := args[0].Val().(string)
	key := args[1].Val().(string)
	myHSet.RLock()
	value, ok := myHSet.data[hash][key]
	myHSet.RUnlock()
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
	myHSet.Lock()
	if myHSet.data[hash] == nil {
		myHSet.data[hash] = make(map[string]any)
	}
	myHSet.data[hash][key] = value
	myHSet.Unlock()
	return resp.NewValue(resp.TAG_STR, "OK")
}
func hgetall(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewValue(resp.TAG_ERR, "Expected exactly 1 argument for HGETALL command")
	}
	hash := args[0].Val().(string)
	myHSet.RLock()
	value := myHSet.data[hash]
	myHSet.RUnlock()

	var arrayValue []resp.Value
	for k, v := range value {
		arrayValue = append(arrayValue, resp.NewValue(resp.TAG_STR, k))
		arrayValue = append(arrayValue, resp.NewValue(resp.TAG_STR, v))
	}
	return resp.NewValue(resp.TAG_ARR, arrayValue)
}

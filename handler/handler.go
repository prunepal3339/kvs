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

	"PUBLISH": publish,
	"PUBSUB":  pubsub,
}
var WriteHandlers = map[string]func(*resp.Writer, []resp.Value) resp.Value{
	"SUBSCRIBE": subscribe,
}

type SafeMap struct {
	sync.RWMutex
	data map[string]any
}
type SafeHMap struct {
	sync.RWMutex
	data map[string]map[string]any
}
type Subscriber struct {
	Ch chan resp.Value
}

type Topic struct {
	sync.RWMutex
	Name        string
	subscribers []*Subscriber
}

func (t *Topic) update(topic resp.Value) {
	if len(t.subscribers) == 0 {
		return
	}
	for _, sub := range t.subscribers {
		sub.Ch <- topic
	}
}

type TopicRegistry struct {
	sync.RWMutex
	data map[string]*Topic
}

func NewTopicRegistry() *TopicRegistry {
	return &TopicRegistry{
		data: make(map[string]*Topic),
	}
}

var topicRegistry = NewTopicRegistry()

func (r *TopicRegistry) getOrCreate(name string) *Topic {
	r.RLock()
	if t, exists := r.data[name]; exists {
		return t
	}
	r.RUnlock()

	r.Lock()
	t := &Topic{
		Name: name,
	}
	r.data[name] = t
	r.Unlock()
	return t
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

func publish(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewValue(resp.TAG_ERR, "[USAGE]: PUBLISH topic message")
	}
	topicName := args[0].Val().(string)
	value := args[1]

	topic := topicRegistry.getOrCreate(topicName)

	topic.Lock()
	topic.update(value)
	topic.Unlock()

	topic.RLock()
	count := len(topic.subscribers)
	topic.RUnlock()

	return resp.NewValue(resp.TAG_INT, count)
}
func subscribe(writer *resp.Writer, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewValue(resp.TAG_ERR, "[USAGE]: SUBSCRIBE topic")
	}
	topicName := args[0].Val().(string)
	topic := topicRegistry.getOrCreate(topicName)
	sub := &Subscriber{
		Ch: make(chan resp.Value),
	}
	topic.Lock()
	topic.subscribers = append(topic.subscribers, sub)
	topic.Unlock()

	for msg := range sub.Ch {
		writer.Write(msg)
	}
	return resp.NewValue(resp.TAG_STR, "SUCCESS")
}

func pubsub(args []resp.Value) resp.Value {
	return resp.NewValue(resp.TAG_STR, "OK")
}

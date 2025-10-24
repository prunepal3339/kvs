package persistence

import (
	"bufio"
	"errors"
	"io"
	"os"
	"time"

	"github.com/prunepal3339/kvs/resp"
)

type Aof struct {
	file   *os.File
	reader *bufio.Reader
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		_ = f
		return nil, err
	}

	aof := &Aof{
		file:   f,
		reader: bufio.NewReader(f),
	}

	//sync file every second.
	go func() {
		for {
			aof.file.Sync() // forces flush buffer to disk.
			time.Sleep(time.Second)
		}
	}()
	return aof, nil
}
func (aof *Aof) Close() error {
	return aof.file.Close()
}
func (aof *Aof) Write(value resp.Value) error {
	_, err := aof.file.Write(value.Marshal())
	return err
}
func (aof *Aof) Read(callback func(value resp.Value)) error {
	reader := resp.NewResp(aof.file)
	for {
		value, err := reader.Read()
		if err != nil {

			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		callback(value)
	}
	return nil
}

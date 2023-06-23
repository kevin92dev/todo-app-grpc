package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"os"
	todo "todo/todo"
)

func main() {
	srv := grpc.NewServer()

	var tasks taskServer
	todo.RegisterTasksServer(srv, tasks)

	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("could not listen port")
	}

	log.Fatal(srv.Serve(l))
}

type taskServer struct{}

const dbPath = "mydb.pb"

func (t taskServer) Add(ctx context.Context, text *todo.Text) (*todo.Void, error) {
	task := &todo.Task{
		Text: text.Text,
		Done: false,
	}

	var void todo.Void
	b, err := proto.Marshal(task)
	if err != nil {
		return &void, fmt.Errorf("Could not encode task: %v", err)
	}

	f, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return &void, fmt.Errorf("Could not open the file %s: %v", dbPath, err)
	}

	if err = gob.NewEncoder(f).Encode(int64(len(b))); err != nil {
		return &void, fmt.Errorf("Could not encode length of message: %v", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return &void, fmt.Errorf("Could not write task on the file: %v", err)
	}

	if err := f.Close(); err != nil {
		return &void, fmt.Errorf("Could not close the file %s: %v", dbPath, err)
	}

	return &void, nil
}

func (t taskServer) List(ctx context.Context, void *todo.Void) (*todo.TaskList, error) {
	b, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read the file %s: %v", dbPath, err)
	}

	var tasks todo.TaskList
	for {
		if len(b) == 0 {
			return &tasks, nil
		} else if len(b) < 4 {
			return nil, fmt.Errorf("Remaining odd %d bytes", len(b))
		}

		var length int64
		if err := gob.NewDecoder(bytes.NewReader(b[:4])).Decode(&length); err != nil {
			return nil, fmt.Errorf("Could not decode message len %d", length)
		}

		b = b[4:]

		var task todo.Task
		if err := proto.Unmarshal(b[:length], &task); err == io.EOF {
			return &todo.TaskList{}, nil
		} else if err != nil {
			return nil, fmt.Errorf("Could not read task: %v", err)
		}

		b = b[length:]

		tasks.Tasks = append(tasks.Tasks, &task)
	}
}

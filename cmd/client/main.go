package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
	todo "todo/todo"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println(os.Stderr, "Missing subcommand: list or add")
		os.Exit(1)
	}

	conn, err := grpc.Dial(":8888", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to backend: %v", err)
	}
	client := todo.NewTasksClient(conn)

	switch cmd := flag.Arg(0); cmd {
	case "list":
		err = list(context.Background(), client)
	case "add":
		err = add(context.Background(), client, strings.Join(flag.Args()[1:], " "))
	default:
		err = fmt.Errorf("Unkown subcommand %s", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func add(ctx context.Context, client todo.TasksClient, text string) error {
	_, err := client.Add(ctx, &todo.Text{Text: text})
	if err != nil {
		return fmt.Errorf("Could not add task: %v", err)
	}

	return nil
}

func list(ctx context.Context, client todo.TasksClient) error {
	l, err := client.List(ctx, &todo.Void{})
	if err != nil {
		return fmt.Errorf("Could not fetch tasks: %v", err)
	}

	for _, t := range l.Tasks {
		if t.Done {
			fmt.Printf("[V] ")
		} else {
			fmt.Printf("[X] ")
		}

		fmt.Println(t.Text)
	}

	return nil
}

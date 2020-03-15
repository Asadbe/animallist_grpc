package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb "github.com/Asadbe/contacts_grpc/task-service/proto/task"
	"google.golang.org/grpc"
)

const (
	address         = "localhost:5431"
	defaultFilename = "task.json"
)

func parseFile(file string) (*pb.Task, error) {
	var task *pb.Task
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}
	return task, err
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect %v", err)
	}

	defer conn.Close()

	client := pb.NewManagingServiceClient(conn)

	file := defaultFilename
	if len(os.Args) > 1 {
		file = os.Args[1]
	}

	task, err := parseFile(file)
	if err != nil {
		log.Fatalf("Could not parse file: %v", err)
	}
	r, err := client.CreateTask(context.Background(), task)
	if err != nil {
		log.Fatalf("Could not greet: %v", err)
	}
	log.Printf("Created: %t", r.Flag)

	
	
	res, err := client.DeleteTask(context.Background(), &pb.DeleteTaskRequest{Id: 4})
	if err != nil {
		log.Fatalf("Could not list tasks: %v", err)
	}
	fmt.Println(res)

	tasks, err := client.SearchTask(context.Background(), &pb.SearchTaskRequest{
		Id: "",
	})
	if err != nil {
		log.Fatalf("Could not list tasks: %v", err)
	}
	fmt.Println(tasks)

	res, err = client.UpdateTask(context.Background(), &pb.UpdateTaskRequest{Id : 4, Task: &pb.Task { 
		Name: "Oybek",
		Voice: "sdf",
		Leg:"4",
	}})
	if err != nil {
		log.Fatalf("Could not list tasks: %v", err)
	}
	fmt.Println(res)
	

	getAll, err := client.GetAllTasks(context.Background(), &pb.GetAllRequest{})
	if err != nil {
		log.Fatalf("Could not list tasks: %v", err)
	}

	for _, v := range getAll.Tasks {
		log.Println(v)
	}
	
}

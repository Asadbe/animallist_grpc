package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	pb "github.com/Asadbe/contacts_grpc/task-service/proto/task"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	host     = "localhost"
	port     = 5431
	user     = "asadbek"
	password = "1"
	dbname   = "mydb"
)

var err error

// Task ...
type Task struct {
	id    int    `db:"serial not null"`
	name  string `db:"not null"`
	voice string `db:"not null"`
	leg   string `db:"not null"`
}

// TaskManagerI ...
type TaskManagerI interface {
	Add(*pb.Task) error
	Update(int64, *pb.Task) error
	Search(name string) ([]*pb.Task,error)
	Delete(id int64) error
	GetAll() ([]*pb.Task, error)
}

// TaskManager ...
type TaskManager struct {
	connectDB *sqlx.DB
}
type sqlxDB struct {
	connectDB *sqlx.DB
}

// NewTaskManager ...
func NewTaskManager() (TaskManagerI, error) {
	cm := sqlxDB{}
	psqlInfo := fmt.Sprintf(`user=%s dbname=%s password=%s`, user, dbname, password)
	cm.connectDB, err = sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &cm, nil
}

// ProtoToStruct ...
func ProtoToStruct(tsk *pb.Task) Task {
	var task Task
	task.name = tsk.Name
	task.voice = tsk.Voice
	task.leg = tsk.Leg
	return task
}

func (s *sqlxDB) Add(a *pb.Task) error {
	insertionQuery := `insert into animals (name, voice, leg) values ($1, $2, $3)`

	_, err := s.connectDB.Exec(insertionQuery, a.Name, a.Voice, a.Leg)

	if err != nil {
		return err
	}

	return nil
}

func (s *sqlxDB) Update(id int64, pb *pb.Task) error {
	updatingQuery := `update animals set name=$1,voice=$2, leg=$3
	where id =$4`

	_, err := s.connectDB.Exec(updatingQuery, pb.Name, pb.Voice, pb.Leg, id)

	if err != nil {
		fmt.Println("Can't update")
		return err
	}

	return nil
}
func (s *sqlxDB) Search(name string) ([]*pb.Task,error) {
	as := []*pb.Task{}
	updatingQuery := `select * from animals where name ilike 'ul%'`

	rows, err := s.connectDB.Query(updatingQuery)
	for rows.Next() {
		ts := &pb.Task{}
		err = rows.Scan(&ts.Id, &ts.Name, &ts.Voice, &ts.Leg)
		if err != nil {
			fmt.Println("Can't scan struct")
			return nil, err
		}
		as=append(as,ts)
	}
		
	if err != nil {
		fmt.Println("Can't update")
		return nil,err
	}

	return as,nil
}

func (s *sqlxDB) Delete(id int64) error {
	fmt.Println("kevotti", id)
	deletingQuery := `delete from animals where id=$1;`

	_, err = s.connectDB.Exec(deletingQuery, id)

	if err != nil {
		fmt.Println("Can't delete")
		return err
	}
	return nil
}

func (s *sqlxDB) GetAll() ([]*pb.Task, error) {
	var (
		tss []*pb.Task
	)

	tss = []*pb.Task{}

	listTaskQuery := `select id,name,voice,leg from animals  `

	rows, err := s.connectDB.Queryx(listTaskQuery)

	if err != nil {
		fmt.Println("Can't print task list")
		return nil, err
	}

	for rows.Next() {
		ts := &pb.Task{}
		err = rows.Scan(&ts.Id, &ts.Name, &ts.Voice, &ts.Leg)
		if err != nil {
			fmt.Println("Can't scan struct")
			return nil, err
		}

		tss = append(tss, ts)
	}
	return tss, nil
}

type service struct {
	tmi TaskManagerI
}

func (s *service) CreateTask(ctx context.Context, req *pb.Task) (*pb.FlagResponse, error) {
	err := s.tmi.Add(req)
	if err != nil {
		return nil, err
	}
	return &pb.FlagResponse{Flag: true}, err
}

func (s *service) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.FlagResponse, error) {
	err := s.tmi.Update(req.Id, req.Task)
	if err != nil {
		return nil, err
	}
	return &pb.FlagResponse{Flag: true}, nil
}

func (s *service) SearchTask(ctx context.Context, req *pb.SearchTaskRequest) (*pb.SearchTaskResponse, error) {
	fmt.Println(req.GetTask().GetName())
	tasks, err := s.tmi.Search(req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.SearchTaskResponse{Tasks: tasks}, nil
}
func (s *service) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.FlagResponse, error) {
	err := s.tmi.Delete(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.FlagResponse{Flag: true}, err
}

func (s *service) GetAllTasks(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	tasks, err := s.tmi.GetAll()
	if err != nil {
		return nil, err
	}
	return &pb.GetAllResponse{Tasks: tasks}, nil
}

func main() {
	tm, err := NewTaskManager()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterManagingServiceServer(s, &service{tm})

	reflection.Register(s)

	log.Println("Running on port:", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

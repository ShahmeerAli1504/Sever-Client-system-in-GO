package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

// Define Matrix and Task structures
type Matrix struct {
	Data [][]int
}

type Task struct {
	Operation string
	MatrixA   Matrix
	MatrixB   Matrix
	ResultA   Matrix // Stores the first matrix result
	ResultB   Matrix // Stores the second matrix result (if needed)
}

// Worker struct
type Worker struct {
	ID      int
	Address string
}

// ProcessTask: Handles matrix operations
func (w *Worker) ProcessTask(task Task, reply *Task) error {
	fmt.Println("Worker received task:", task.Operation)

	// Simulate a delay of 10 seconds before processing the task
	time.Sleep(10 * time.Second)

	switch task.Operation {
	case "add":
		reply.ResultA.Data = AddMatrices(task.MatrixA.Data, task.MatrixB.Data)
	case "multiply":
		reply.ResultA.Data = MultiplyMatrices(task.MatrixA.Data, task.MatrixB.Data)
	case "transpose":
		// Transpose MatrixA
		reply.ResultA.Data = TransposeMatrix(task.MatrixA.Data)

		// Transpose MatrixB (if present)
		if len(task.MatrixB.Data) > 0 {
			reply.ResultB.Data = TransposeMatrix(task.MatrixB.Data)
		}
	default:
		return fmt.Errorf("unknown operation: %s", task.Operation)
	}

	// Log results
	fmt.Println("Processed Result A:", reply.ResultA.Data)
	if len(reply.ResultB.Data) > 0 {
		fmt.Println("Processed Result B:", reply.ResultB.Data)
	}
	return nil
}

// StartWorker: Registers worker with the coordinator
func StartWorker(coordinatorAddr string) {
	rand.Seed(time.Now().UnixNano()) // Ensure randomness in Worker IDs
	worker := &Worker{
		ID: rand.Intn(1000000), // Generate a unique worker ID
	}

	// Find an available port dynamically
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Auto-select free port
	if err != nil {
		fmt.Println("Error finding free port:", err)
		return
	}
	defer listener.Close()

	worker.Address = listener.Addr().String() // Get assigned port

	// Start RPC server
	server := rpc.NewServer()
	server.Register(worker)

	// Register worker with coordinator
	client, err := rpc.Dial("tcp", coordinatorAddr)
	if err != nil {
		fmt.Println("Failed to connect to coordinator:", err)
		return
	}

	var reply string
	err = client.Call("Coordinator.RegisterWorker", worker, &reply)
	if err != nil {
		fmt.Println("Worker registration failed:", err)
	} else {
		fmt.Println(reply)
	}

	// Send heartbeat every 10 seconds
	go func() {
		for {
			time.Sleep(10 * time.Second)
			var reply string
			err := worker.Heartbeat("localhost:1234", &reply)
			if err != nil {
				fmt.Println("Error sending heartbeat:", err)
			} else {
				fmt.Println(reply)
			}
		}
	}()

	// Start serving requests
	fmt.Printf("Worker %d running on %s...\n", worker.ID, worker.Address)
	for {
		conn, _ := listener.Accept()
		go server.ServeConn(conn)
	}
}

// Heartbeat sends a signal to the coordinator that the worker is alive
func (w *Worker) Heartbeat(coordinatorAddr string, reply *string) error {
	client, err := rpc.Dial("tcp", coordinatorAddr)
	if err != nil {
		return fmt.Errorf("unable to connect to coordinator: %v", err)
	}

	err = client.Call("Coordinator.ReceiveHeartbeat", w.ID, reply)
	if err != nil {
		return fmt.Errorf("heartbeat failed: %v", err)
	}

	return nil
}

// Helper functions for matrix operations
func AddMatrices(a, b [][]int) [][]int {
	rows, cols := len(a), len(a[0])
	result := make([][]int, rows)
	for i := range result {
		result[i] = make([]int, cols)
		for j := range result[i] {
			result[i][j] = a[i][j] + b[i][j]
		}
	}
	return result
}

func TransposeMatrix(a [][]int) [][]int {
	rows, cols := len(a), len(a[0])
	result := make([][]int, cols)
	for i := range result {
		result[i] = make([]int, rows)
		for j := range result[i] {
			result[i][j] = a[j][i]
		}
	}
	return result
}

func MultiplyMatrices(a, b [][]int) [][]int {
	rowsA, colsA, colsB := len(a), len(a[0]), len(b[0])
	result := make([][]int, rowsA)
	for i := range result {
		result[i] = make([]int, colsB)
		for j := range result[i] {
			for k := 0; k < colsA; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

func main() {
	StartWorker("localhost:1234") // Replace with Coordinator IP
}

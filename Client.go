package main

import (
	"fmt"
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

func main() {
	// Dial the server
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error dialing server:", err)
		return
	}
	defer client.Close()

	matrixA := Matrix{Data: [][]int{{1, 2}, {3, 4}}}
	matrixB := Matrix{Data: [][]int{{5, 6}, {7, 8}}}
	task := Task{Operation: "transpose", MatrixA: matrixA, MatrixB: matrixB}

	var result Task

	// Send the task to the Coordinator
	fmt.Println("Forwarding task to Coordinator:", task.Operation)
	err = client.Call("Coordinator.AssignTask", task, &result)
	if err != nil {
		fmt.Println("Task failed:", err)
		return
	}

	// Wait for 30 seconds before retrieving the result
	fmt.Println("Task forwarded successfully. Waiting for 30 seconds before retrieving the result...")
	time.Sleep(30 * time.Second)

	// Display the result after waiting
	fmt.Println("Retrieving the result now...")

	switch task.Operation {
	case "transpose":
		fmt.Println("Transposed Matrix A:")
		printMatrix(result.ResultA.Data)
		if len(result.ResultB.Data) > 0 {
			fmt.Println("Transposed Matrix B:")
			printMatrix(result.ResultB.Data)
		}
	default: // For "multiply" and "add"
		fmt.Println("Result:")
		printMatrix(result.ResultA.Data)
	}
}

// Utility function to print matrices
func printMatrix(matrix [][]int) {
	for _, row := range matrix {
		fmt.Println(row)
	}
}

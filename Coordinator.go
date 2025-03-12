package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
)

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

type Worker struct {
	ID      int
	Address string
	Busy    bool
}

type WorkerStatus struct {
	LastSeen time.Time
}

type Coordinator struct {
	mu           sync.Mutex
	workers      map[int]*Worker
	workerStatus map[int]*WorkerStatus // Track worker status
	tasks        chan Task
}

// RegisterWorker: Adds a worker to the pool
func (c *Coordinator) RegisterWorker(worker Worker, reply *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure the worker ID is unique
	if _, exists := c.workers[worker.ID]; exists {
		worker.ID = rand.Intn(1000000) // Generate a new unique ID if conflict happens
	}

	c.workers[worker.ID] = &worker
	*reply = fmt.Sprintf("Worker %d registered at %s", worker.ID, worker.Address)
	fmt.Println(*reply)
	return nil
}

// AssignTask: Sends a task to an available worker or queues it
func (c *Coordinator) AssignTask(task Task, reply *Task) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Log when a client connects and sends a task
	fmt.Println("\n[Client Connected] Received Task:", task.Operation)

	// Find an available worker and try to send the task
	for _, worker := range c.workers {
		if !worker.Busy {
			worker.Busy = true
			fmt.Printf("[Task Assignment] Task '%s' assigned to Worker %d at %s\n", task.Operation, worker.ID, worker.Address)

			err := c.sendTaskToWorker(worker, task, reply)
			worker.Busy = false
			if err != nil {
				// Try next worker in case of failure
				fmt.Println("[Error] Worker failed, retrying task with another worker...")
				continue
			}
			return nil
		}
	}

	// If no worker is available, queue the task
	fmt.Println("[Task Queued] No available workers, task added to queue.")
	c.tasks <- task
	return nil
}

// sendTaskToWorker: Calls the worker to process a task, with retry on failure
func (c *Coordinator) sendTaskToWorker(worker *Worker, task Task, reply *Task) error {
	fmt.Printf("[Task Processing] Worker %d is processing task: %s\n", worker.ID, task.Operation)

	client, err := rpc.Dial("tcp", worker.Address)
	if err != nil {
		fmt.Println("[Error] Worker unreachable, retrying task...")
		worker.Busy = false
		c.tasks <- task
		return err
	}

	// Store result in a temporary variable
	var workerResult Task
	err = client.Call("Worker.ProcessTask", task, &workerResult)
	worker.Busy = false

	if err != nil {
		fmt.Println("[Error] Worker failed task execution. Requeuing task...")
		c.tasks <- task
		return err
	}

	// Copy both results into the reply
	reply.ResultA.Data = workerResult.ResultA.Data
	reply.ResultB.Data = workerResult.ResultB.Data

	// Confirm result
	fmt.Println("[Task Completed] Worker", worker.ID, "completed task:", task.Operation)
	fmt.Println("Result A:", reply.ResultA.Data)
	if len(reply.ResultB.Data) > 0 {
		fmt.Println("Result B:", reply.ResultB.Data)
	}
	return nil
}

// workerDispatcher: Assigns queued tasks when workers become free
func (c *Coordinator) workerDispatcher() {
	for {
		task := <-c.tasks
		c.mu.Lock()
		for _, worker := range c.workers {
			if !worker.Busy {
				worker.Busy = true
				fmt.Printf("[Reassigning Task] Assigning queued task '%s' to Worker %d at %s\n", task.Operation, worker.ID, worker.Address)
				go c.sendTaskToWorker(worker, task, &Task{}) // Fixed argument type
				break
			}
		}
		c.mu.Unlock()
	}
}

// ReceiveHeartbeat: Receives heartbeat signals from workers
func (c *Coordinator) ReceiveHeartbeat(workerID int, reply *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update the last seen time for the worker
	if status, exists := c.workerStatus[workerID]; exists {
		status.LastSeen = time.Now()
	} else {
		c.workerStatus[workerID] = &WorkerStatus{LastSeen: time.Now()}
	}

	*reply = fmt.Sprintf("Received heartbeat from Worker %d", workerID)
	return nil
}

// monitorWorkers: Checks worker heartbeats and reassigns tasks if necessary
func (c *Coordinator) monitorWorkers() {
	for {
		c.mu.Lock()
		for workerID, status := range c.workerStatus {
			// If no heartbeat in the last 30 seconds, consider the worker as crashed
			if time.Since(status.LastSeen) > 30*time.Second {
				fmt.Printf("[Warning] Worker %d has not sent heartbeat. Reassigning tasks...\n", workerID)
				// Reassign tasks from this worker
				// (You may need to track which tasks were assigned to the worker)
			}
		}
		c.mu.Unlock()

		time.Sleep(10 * time.Second)
	}
}

func (c *Coordinator) StartCoordinator() {
	c.workers = make(map[int]*Worker)
	c.workerStatus = make(map[int]*WorkerStatus)
	c.tasks = make(chan Task, 100) // Buffer size for task queue

	rpc.Register(c)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("[Error] Could not start Coordinator:", err)
		return
	}
	defer listener.Close()

	fmt.Println("[Coordinator] Running on port 1234...")

	go c.workerDispatcher()
	go c.monitorWorkers()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[Error] Client connection failed:", err)
			continue
		}
		fmt.Println("\n[Client Connected] New connection established.")
		go rpc.ServeConn(conn)
	}
}

func main() {
	coordinator := &Coordinator{}
	coordinator.StartCoordinator()
}

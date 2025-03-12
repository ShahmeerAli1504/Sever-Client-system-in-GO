# Sever-Client-system-in-GO

## Scenario
The client programs will request the coordinator for computation on their data. The server will act as a coordinator for n worker processes (n ≥ 3). 
The worker processes will be responsible for performing matrix operations, limited to:
• Addition
• Transpose
• Multiplication
The coordinator is responsible for assigning tasks to the workers and sending the results back to the client.
### Client’s Role
• The client will initiate computation requests from the coordinator process.
• The client will request services via RPC (Remote Procedure Calls), ensuring that the client and server programs run on different physical devices.
### Coordinator’s Responsibilities
• The coordinator (server) will schedule the tasks on a First-Come, First-Served (FCFS) basis.
• The coordinator will assign tasks to the least busy workers first (Load Balancing).
• In case of a worker’s failure, the server will assign the task to the next available worker (Fault Tolerance).
• The coordinator will gather the results from workers and send them back to the client.
### Worker’s Responsibilities
• The worker will be responsible for performing matrix operations on the data received from the coordinator process.

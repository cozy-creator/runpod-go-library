# RunPod Go Library

A comprehensive Go client library for the RunPod REST API, providing programmatic access to GPU cloud resources, serverless endpoints, and pod management.

## 🚀 Features

- ✅ **Pod Management** - Create, monitor, and manage GPU/CPU pods
- ✅ **Serverless Jobs** - Submit, monitor, and manage serverless job execution
- ✅ **Complete REST API** - Full RunPod REST API support
- ✅ **Error Handling** - Comprehensive error types and retry logic
- ✅ **Type Safety** - Strong typing for all API responses
- ✅ **Debug Support** - Optional request/response logging
- ✅ **Thread Safe** - Safe for concurrent use
- ✅ **Streaming Support** - Real-time job result streaming
- 🔄 **Endpoint Management** - Serverless endpoint lifecycle (coming soon)
- 🔄 **Templates** - Pod and serverless templates (coming soon)

## 📦 Installation

```bash
go mod init your-project
go get github.com/cozy-creator/runpod-go-library
```

## 🎯 Quick Start

### Basic Client Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/cozy-creator/runpod-go-library"
)

func main() {
    // Create client with your RunPod API key
    client := runpod.NewClient("your-runpod-api-key")
    
    // Or with custom options
    client = runpod.NewClient("your-api-key",
        runpod.WithDebug(true),                    // Enable debug logging
        runpod.WithTimeout(60*time.Second),        // Custom timeout
        runpod.WithMaxRetryAttempts(5),           // Retry failed requests
    )
    
    fmt.Printf("✅ Client ready! Base URL: %s\n", client.GetBaseURL())
}
```

### Pod Management

```go
ctx := context.Background()

// 1. Create a pod (simple version)
envVars := map[string]string{
    "JUPYTER_PASSWORD": "secure-password",
    "WORKSPACE_DIR":    "/workspace",
}

podID, err := client.LaunchRunPod(ctx, "runpod/pytorch:2.1.0-py3.10-cuda11.8.0", envVars)
if err != nil {
    log.Fatal("Failed to create pod:", err)
}
fmt.Printf("🎉 Pod created: %s\n", podID)

// 2. Get pod status
status, err := client.GetPodStatus(ctx, podID)
if err != nil {
    log.Fatal("Failed to get status:", err)
}
fmt.Printf("📊 Pod status: %s\n", status)

// 3. Get full pod details
pod, err := client.GetPod(ctx, podID)
if err != nil {
    log.Fatal("Failed to get pod:", err)
}
fmt.Printf("💰 Cost per hour: $%.4f\n", pod.CostPerHour)
fmt.Printf("🖥️  GPU: %s\n", pod.GPUTypeID)

// 4. Wait for pod to be running
pod, err = client.WaitForPodStatus(ctx, podID, "RUNNING", 30)
if err != nil {
    log.Fatal("Pod failed to start:", err)
}
fmt.Printf("✅ Pod is now running!\n")

// 5. List all your pods
pods, err := client.ListPods(ctx, &runpod.ListOptions{Limit: 10})
if err != nil {
    log.Fatal("Failed to list pods:", err)
}
fmt.Printf("📋 You have %d pods\n", len(pods))

// 6. Terminate the pod when done
err = client.TerminatePod(ctx, podID)
if err != nil {
    log.Fatal("Failed to terminate pod:", err)
}
fmt.Printf("🗑️  Pod terminated\n")
```

### Serverless Job Management

```go
ctx := context.Background()

// 1. Submit an asynchronous job
input := map[string]interface{}{
    "prompt": "A beautiful sunset over mountains",
    "steps":  20,
    "width":  512,
    "height": 512,
}

job, err := client.RunAsync(ctx, "your-endpoint-id", input)
if err != nil {
    log.Fatal("Failed to submit job:", err)
}
fmt.Printf("🚀 Job submitted: %s (Status: %s)\n", job.ID, job.Status)

// 2. Monitor job progress
for {
    job, err = client.GetJobStatus(ctx, "your-endpoint-id", job.ID)
    if err != nil {
        log.Fatal("Failed to get job status:", err)
    }
    
    fmt.Printf("📊 Job %s status: %s\n", job.ID, job.Status)
    
    if client.IsJobTerminal(job.Status) {
        break
    }
    
    time.Sleep(2 * time.Second)
}

// 3. Get final results
if job.Status == "COMPLETED" {
    fmt.Printf("✅ Job completed! Output: %+v\n", job.Output)
} else {
    fmt.Printf("❌ Job failed: %s\n", job.Error)
}

// 4. Submit synchronous job (wait for completion)
syncJob, err := client.RunSync(ctx, "your-endpoint-id", input)
if err != nil {
    log.Fatal("Failed to run sync job:", err)
}
fmt.Printf("⚡ Sync job completed: %+v\n", syncJob.Output)

// 5. Stream job results in real-time
jobChan, errChan := client.StreamResultsContinuous(ctx, "your-endpoint-id", job.ID, 1*time.Second)

for {
    select {
    case job := <-jobChan:
        if job == nil {
            fmt.Println("🏁 Streaming completed")
            return
        }
        fmt.Printf("📡 Streaming update: %s - %+v\n", job.Status, job.Output)
        
        if client.IsJobTerminal(job.Status) {
            fmt.Println("🏁 Job completed via streaming")
            return
        }
        
    case err := <-errChan:
        log.Printf("❌ Streaming error: %v", err)
        return
        
    case <-time.After(30 * time.Second):
        fmt.Println("⏰ Streaming timeout")
        return
    }
}
```

### Advanced Job Operations

```go
ctx := context.Background()

// 1. Submit multiple jobs in batch
inputs := []interface{}{
    map[string]string{"prompt": "cat"},
    map[string]string{"prompt": "dog"},
    map[string]string{"prompt": "bird"},
}

jobs, err := client.SubmitMultipleJobs(ctx, "your-endpoint-id", inputs)
if err != nil {
    log.Fatal("Failed to submit multiple jobs:", err)
}
fmt.Printf("🔄 Submitted %d jobs\n", len(jobs))

// 2. Run and wait with timeout
job, err := client.RunAndWait(ctx, "your-endpoint-id", input, 5*time.Minute)
if err != nil {
    log.Fatal("Job failed or timed out:", err)
}
fmt.Printf("⏱️  Job completed in %d seconds\n", job.ExecutionTime)

// 3. Quick run (tries sync first, falls back to async)
job, err = client.QuickRun(ctx, "your-endpoint-id", input)
if err != nil {
    log.Fatal("Quick run failed:", err)
}
fmt.Printf("🏃 Quick run result: %+v\n", job.Output)

// 4. Job management operations
err = client.CancelJob(ctx, "your-endpoint-id", "job-id")
if err != nil {
    log.Printf("Failed to cancel job: %v", err)
}

retryJob, err := client.RetryJob(ctx, "your-endpoint-id", "failed-job-id")
if err != nil {
    log.Printf("Failed to retry job: %v", err)
}

err = client.PurgeQueue(ctx, "your-endpoint-id")
if err != nil {
    log.Printf("Failed to purge queue: %v", err)
}

// 5. Check endpoint health
health, err := client.GetHealth(ctx, "your-endpoint-id")
if err != nil {
    log.Fatal("Failed to get health:", err)
}
fmt.Printf("🏥 Endpoint health: %s (Queue: %d, Workers: %d/%d)\n", 
    health.Status, health.JobsInQueue, health.WorkersActive, health.WorkersTotal)
```

### Advanced Pod Creation

```go
// Create pod with full configuration
req := &runpod.CreatePodRequest{
    Name:              "my-training-pod",
    ImageName:         "runpod/pytorch:2.1.0-py3.10-cuda11.8.0",
    GPUTypeID:         "NVIDIA GeForce RTX 4090",
    GPUCount:          1,
    VCPUCount:         4,
    MemoryInGB:        16,
    ContainerDiskInGB: 50,
    VolumeInGB:        100,
    VolumeMountPath:   "/workspace",
    DatacenterID:      "US-CA-1",
    CloudType:         "SECURE",
    Env: map[string]string{
        "JUPYTER_PASSWORD": "secure-password",
        "WANDB_API_KEY":    "your-wandb-key",
    },
    Ports:      []string{"8888/http", "6006/http"},
    DockerArgs: "--shm-size=1g",
}

pod, err := client.CreatePod(ctx, req)
if err != nil {
    log.Fatal("Failed to create pod:", err)
}
fmt.Printf("🎉 Advanced pod created: %s\n", pod.ID)
```

### Spot/Interruptible Pods

```go
// Create a spot instance with bidding
req := &runpod.CreatePodRequest{
    Name:              "spot-training-pod",
    ImageName:         "runpod/pytorch:latest",
    GPUTypeID:         "NVIDIA GeForce RTX 4090",
    GPUCount:          1,
    ContainerDiskInGB: 50,
    BidPerGPU:         0.50, // Bid $0.50 per GPU per hour
    CloudType:         "COMMUNITY",
}

pod, err := client.CreatePod(ctx, req)
// or use the convenience function:
// pod, err := client.CreateSpotPod(ctx, req, 0.50)
```

## 🔧 Configuration Options

```go
client := runpod.NewClient("your-api-key",
    // API Configuration
    runpod.WithBaseURL("https://custom.runpod.io/v1"),     // Custom API URL
    runpod.WithServerlessBaseURL("https://custom.api.runpod.ai/v2"), // Custom serverless URL
    
    // HTTP Configuration  
    runpod.WithTimeout(120*time.Second),                   // Request timeout
    runpod.WithHTTPClient(customHTTPClient),               // Custom HTTP client
    
    // Retry Configuration
    runpod.WithMaxRetryAttempts(5),                        // Max retry attempts
    runpod.WithRetryDelay(2*time.Second),                  // Delay between retries
    
    // Debug Configuration
    runpod.WithDebug(true),                                // Enable debug logging
    runpod.WithLogger(customLogger),                       // Custom logger
    runpod.WithUserAgent("my-app/1.0"),                    // Custom user agent
)
```

## 🛠️ Pod Management Functions

| Function | Description |
|----------|-------------|
| `LaunchRunPod()` | Quick pod creation with defaults |
| `CreatePod()` | Full pod creation with all options |
| `GetPod()` | Get complete pod details |
| `GetPodStatus()` | Get just the pod status |
| `ListPods()` | List all pods with pagination |
| `StopPod()` | Stop a running pod |
| `ResumePod()` | Resume a stopped pod |
| `TerminatePod()` | Terminate/delete a pod |
| `GetPodLogs()` | Get pod logs |
| `WaitForPodStatus()` | Wait for specific status |
| `FindPodByName()` | Find pod by name |

## ⚡ Serverless Job Functions

| Function | Description |
|----------|-------------|
| `RunAsync()` | Submit asynchronous job |
| `RunSync()` | Submit synchronous job |
| `GetJobStatus()` | Get job status and results |
| `WaitForJobCompletion()` | Wait for job to complete |
| `StreamResults()` | Stream job results once |
| `StreamResultsContinuous()` | Stream job results continuously |
| `CancelJob()` | Cancel running job |
| `RetryJob()` | Retry failed job |
| `PurgeQueue()` | Clear endpoint queue |
| `GetHealth()` | Get endpoint health |
| `SubmitMultipleJobs()` | Submit multiple jobs |
| `RunAndWait()` | Submit job and wait for completion |
| `QuickRun()` | Smart job submission (sync/async) |
| `IsJobTerminal()` | Check if job status is final |

## 🚨 Error Handling

The library provides detailed error classification:

```go
ctx := context.Background()
_, err := client.GetPod(ctx, "invalid-pod-id")

if err != nil {
    switch {
    case runpod.IsAPIError(err):
        apiErr := err.(*runpod.APIError)
        if apiErr.IsNotFound() {
            fmt.Println("Pod not found")
        } else if apiErr.IsUnauthorized() {
            fmt.Println("Invalid API key")
        } else if apiErr.IsRateLimited() {
            fmt.Println("Rate limited")
        }
        
    case runpod.IsNetworkError(err):
        fmt.Println("Network connectivity issue")
        
    case runpod.IsTimeoutError(err):
        fmt.Println("Request timed out")
        
    case runpod.IsValidationError(err):
        fmt.Println("Invalid input parameters")
    }
}
```

### Available Error Types

- **`APIError`** - HTTP errors from RunPod API (4xx, 5xx)
- **`ValidationError`** - Input validation errors  
- **`NetworkError`** - Network connectivity issues
- **`TimeoutError`** - Request timeout errors
- **`AuthError`** - Authentication/authorization errors
- **`RateLimitError`** - Rate limiting errors

## 🔍 Debug Mode

Enable debug mode to see detailed request/response information:

```go
client := runpod.NewClient("your-api-key", runpod.WithDebug(true))

// This will output:
// [DEBUG] POST https://rest.runpod.io/v1/pods
// [DEBUG] Request Body: {"name": "test-pod", "imageName": "runpod/pytorch", ...}
// [DEBUG] Response Status: 200
// [DEBUG] Response Body: {"id": "pod-123", "status": "CREATED", ...}
```

## 📊 Type Definitions

All RunPod API objects are strongly typed:

```go
type Pod struct {
    ID                string            `json:"id"`
    Name              string            `json:"name"`  
    Status            string            `json:"status"`
    ImageName         string            `json:"imageName"`
    GPUCount          int               `json:"gpuCount"`
    GPUTypeID         string            `json:"gpuTypeId"`
    CostPerHour       float64           `json:"costPerHr"`
    CreatedAt         time.Time         `json:"createdAt"`
    Env               map[string]string `json:"env"`
    // ... and many more fields
}

type Job struct {
    ID            string                 `json:"id"`
    Status        string                 `json:"status"`
    Input         map[string]interface{} `json:"input"`
    Output        interface{}            `json:"output"`
    Error         string                 `json:"error"`
    CreatedAt     *JSONTime              `json:"createdAt"`
    StartedAt     *JSONTime              `json:"startedAt"`
    CompletedAt   *JSONTime              `json:"completedAt"`
    ExecutionTime int                    `json:"executionTimeInMs"`
    EndpointID    string                 `json:"endpointId"`
}

type EndpointHealth struct {
    Status        string `json:"status"`
    JobsInQueue   int    `json:"jobsInQueue"`
    WorkersIdle   int    `json:"workersIdle"`
    WorkersActive int    `json:"workersRunning"`
    WorkersTotal  int    `json:"workersTotal"`
}
```

## 🎯 What's Implemented Now

- ✅ **Phase 1: Core Infrastructure** - Client, authentication, error handling
- ✅ **Phase 2: Pod Management** - Complete pod lifecycle management  
- ✅ **Phase 3: Serverless Jobs** - Complete job execution and monitoring

## 🚧 Coming Soon

### Phase 4: Endpoint Management 🔄  
- [ ] **CreateEndpoint** - Create new serverless endpoints
- [ ] **GetEndpoint** - Get endpoint details and configuration
- [ ] **ListEndpoints** - List all your serverless endpoints
- [ ] **UpdateEndpoint** - Update endpoint configuration
- [ ] **DeleteEndpoint** - Delete serverless endpoints

### Phase 5: Templates 📄
- [ ] **CreateTemplate** - Create pod and serverless templates
- [ ] **GetTemplate** - Get template details
- [ ] **ListTemplates** - List available templates
- [ ] **UpdateTemplate** - Update template configuration  
- [ ] **DeleteTemplate** - Delete templates

### Phase 6: Resource Information 📊
- [ ] **ListGPUTypes** - Get available GPU types and pricing
- [ ] **GetGPUPricing** - Get current GPU pricing information
- [ ] **ListDatacenters** - Get available datacenter locations
- [ ] **GetAccountInfo** - Get account details and limits
- [ ] **GetUsageStats** - Get usage statistics and billing info

### Phase 7: Advanced Features 🔧
- [ ] **WebhookConfiguration** - Configure webhooks for job completion
- [ ] **BulkOperations** - Batch operations for multiple pods/jobs
- [ ] **FileUpload/Download** - Handle large file transfers
- [ ] **NetworkVolumes** - Manage persistent storage volumes
- [ ] **Secrets Management** - Handle environment secrets securely

## 🧪 Testing

The library includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test suites
go test -v ./tests/ -run TestPod
go test -v ./tests/ -run TestJob
go test -v ./tests/ -run TestStream
```

### Test Features
- ✅ **Unit tests** with comprehensive mock servers
- ✅ **Integration tests** for real API validation
- ✅ **Error handling tests** for all error types
- ✅ **Streaming tests** for real-time job monitoring
- ✅ **Concurrent safety tests** for thread safety

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## 📝 License

This project is licensed under the MIT License.

## 🔗 Links

- [RunPod Platform](https://runpod.io)
- [RunPod API Documentation](https://docs.runpod.io)
- [RunPod REST API Reference](https://rest.runpod.io/v1/docs)
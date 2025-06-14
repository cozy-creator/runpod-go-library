package runpod_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Configuration
const (
	// Test configuration
	NumIterations = 5
	TestImage     = "runpod/pytorch:2.1.0-py3.10-cuda11.8.0-devel-ubuntu22.04"
	TestGPUType   = "NVIDIA GeForce RTX 4090"
)

// GraphQL API structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type GraphQLPodResponse struct {
	Data struct {
		PodFindAndDeployOnDemand struct {
			ID            string `json:"id"`
			ImageName     string `json:"imageName"`
			MachineID     string `json:"machineId"`
			DesiredStatus string `json:"desiredStatus"`
		} `json:"podFindAndDeployOnDemand"`
	} `json:"data"`
}

type GraphQLTerminateResponse struct {
	Data struct {
		PodTerminate interface{} `json:"podTerminate"`
	} `json:"data"`
}

// REST API structures (from your library)
type RESTCreatePodRequest struct {
	Name              string            `json:"name"`
	ImageName         string            `json:"imageName"`
	GPUTypeIDs        []string          `json:"gpuTypeIds"`
	GPUCount          int               `json:"gpuCount"`
	ContainerDiskInGB int               `json:"containerDiskInGb"`
	VolumeInGB        int               `json:"volumeInGb"`
	VCPUCount         int               `json:"vcpuCount"`
	CloudType         string            `json:"cloudType"`
	ComputeType       string            `json:"computeType"`
	SupportPublicIP   bool              `json:"supportPublicIp"`
	Env               map[string]string `json:"env,omitempty"`
	Ports             []string          `json:"ports,omitempty"`
}

type RESTPodResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	DesiredStatus string            `json:"desiredStatus"`
	ImageName     string            `json:"image"`
	MachineID     string            `json:"machineId"`
	Env           map[string]string `json:"env"`
}

// Test result structures
type LatencyResult struct {
	CreateTime    time.Duration
	TerminateTime time.Duration
	TotalTime     time.Duration
	Success       bool
	Error         string
}

type ComparisonResults struct {
	GraphQLResults []LatencyResult
	RESTResults    []LatencyResult
	GraphQLAvg     LatencyResult
	RESTAvg        LatencyResult
}

// API clients
type GraphQLClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

type RESTClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func NewGraphQLClient(apiKey string) *GraphQLClient {
	return &GraphQLClient{
		APIKey:  apiKey,
		BaseURL: "https://api.runpod.io/graphql",
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func NewRESTClient(apiKey string) *RESTClient {
	return &RESTClient{
		APIKey:  apiKey,
		BaseURL: "https://rest.runpod.io/v1",
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GraphQL Methods
func (c *GraphQLClient) CreatePod(ctx context.Context, name string) (*GraphQLPodResponse, error) {
	query := `
	mutation {
		podFindAndDeployOnDemand(input: {
			cloudType: SECURE
			gpuCount: 1
			gpuTypeId: "` + TestGPUType + `"
			name: "` + name + `"
			imageName: "` + TestImage + `"
			containerDiskInGb: 50
			volumeInGb: 20
			minVcpuCount: 2
			minMemoryInGb: 15
			supportPublicIp: true
		}) {
			id
			imageName
			machineId
			desiredStatus
		}
	}`

	req := GraphQLRequest{Query: query}

	var response GraphQLPodResponse
	err := c.makeRequest(ctx, req, &response)
	return &response, err
}

func (c *GraphQLClient) TerminatePod(ctx context.Context, podID string) error {
	query := `
	mutation {
		podTerminate(input: {
			podId: "` + podID + `"
		})
	}`

	req := GraphQLRequest{Query: query}

	var response GraphQLTerminateResponse
	return c.makeRequest(ctx, req, &response)
}

func (c *GraphQLClient) makeRequest(ctx context.Context, req GraphQLRequest, result interface{}) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"?api_key="+c.APIKey, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Check for GraphQL errors
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return err
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return json.Unmarshal(body, result)
}

// REST Methods
func (c *RESTClient) CreatePod(ctx context.Context, name string) (*RESTPodResponse, error) {
	req := RESTCreatePodRequest{
		Name:              name,
		ImageName:         TestImage,
		GPUTypeIDs:        []string{TestGPUType},
		GPUCount:          1,
		ContainerDiskInGB: 50,
		VolumeInGB:        20,
		VCPUCount:         2,
		CloudType:         "SECURE",
		ComputeType:       "GPU",
		SupportPublicIP:   true,
	}

	var response RESTPodResponse
	err := c.makeRequest(ctx, "POST", "/pods", req, &response)
	return &response, err
}

func (c *RESTClient) TerminatePod(ctx context.Context, podID string) error {
	return c.makeRequest(ctx, "DELETE", "/pods/"+podID, nil, nil)
}

func (c *RESTClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}

	return nil
}

// Test execution functions
func testGraphQLLatency(client *GraphQLClient, iteration int) LatencyResult {
	ctx := context.Background()
	name := fmt.Sprintf("gql-latency-test-%d-%d", time.Now().Unix(), iteration)

	// Measure pod creation
	createStart := time.Now()
	pod, err := client.CreatePod(ctx, name)
	createTime := time.Since(createStart)

	if err != nil {
		return LatencyResult{
			CreateTime: createTime,
			Success:    false,
			Error:      fmt.Sprintf("Create failed: %v", err),
		}
	}

	podID := pod.Data.PodFindAndDeployOnDemand.ID
	if podID == "" {
		return LatencyResult{
			CreateTime: createTime,
			Success:    false,
			Error:      "No pod ID returned",
		}
	}

	log.Printf("GraphQL iteration %d: Created pod %s in %v", iteration, podID, createTime)

	// Wait a moment before terminating
	time.Sleep(2 * time.Second)

	// Measure pod termination
	terminateStart := time.Now()
	err = client.TerminatePod(ctx, podID)
	terminateTime := time.Since(terminateStart)

	if err != nil {
		return LatencyResult{
			CreateTime:    createTime,
			TerminateTime: terminateTime,
			Success:       false,
			Error:         fmt.Sprintf("Terminate failed: %v", err),
		}
	}

	log.Printf("GraphQL iteration %d: Terminated pod %s in %v", iteration, podID, terminateTime)

	return LatencyResult{
		CreateTime:    createTime,
		TerminateTime: terminateTime,
		TotalTime:     createTime + terminateTime,
		Success:       true,
	}
}

func testRESTLatency(client *RESTClient, iteration int) LatencyResult {
	ctx := context.Background()
	name := fmt.Sprintf("rest-latency-test-%d-%d", time.Now().Unix(), iteration)

	// Measure pod creation
	createStart := time.Now()
	pod, err := client.CreatePod(ctx, name)
	createTime := time.Since(createStart)

	if err != nil {
		return LatencyResult{
			CreateTime: createTime,
			Success:    false,
			Error:      fmt.Sprintf("Create failed: %v", err),
		}
	}

	podID := pod.ID
	if podID == "" {
		return LatencyResult{
			CreateTime: createTime,
			Success:    false,
			Error:      "No pod ID returned",
		}
	}

	log.Printf("REST iteration %d: Created pod %s in %v", iteration, podID, createTime)

	// Wait a moment before terminating
	time.Sleep(2 * time.Second)

	// Measure pod termination
	terminateStart := time.Now()
	err = client.TerminatePod(ctx, podID)
	terminateTime := time.Since(terminateStart)

	if err != nil {
		return LatencyResult{
			CreateTime:    createTime,
			TerminateTime: terminateTime,
			Success:       false,
			Error:         fmt.Sprintf("Terminate failed: %v", err),
		}
	}

	log.Printf("REST iteration %d: Terminated pod %s in %v", iteration, podID, terminateTime)

	return LatencyResult{
		CreateTime:    createTime,
		TerminateTime: terminateTime,
		TotalTime:     createTime + terminateTime,
		Success:       true,
	}
}

func calculateAverage(results []LatencyResult) LatencyResult {
	if len(results) == 0 {
		return LatencyResult{}
	}

	var totalCreate, totalTerminate, totalTime time.Duration
	successCount := 0

	for _, result := range results {
		if result.Success {
			totalCreate += result.CreateTime
			totalTerminate += result.TerminateTime
			totalTime += result.TotalTime
			successCount++
		}
	}

	if successCount == 0 {
		return LatencyResult{Success: false, Error: "No successful runs"}
	}

	return LatencyResult{
		CreateTime:    totalCreate / time.Duration(successCount),
		TerminateTime: totalTerminate / time.Duration(successCount),
		TotalTime:     totalTime / time.Duration(successCount),
		Success:       true,
	}
}

func printResults(results ComparisonResults) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🚀 RUNPOD API LATENCY COMPARISON RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	// Individual results
	fmt.Println("\n📊 INDIVIDUAL TEST RESULTS:")
	fmt.Println(strings.Repeat("-", 70))

	fmt.Printf("%-15s %-15s %-15s %-15s %s\n", "Test", "Create", "Terminate", "Total", "Status")
	fmt.Println(strings.Repeat("-", 70))

	for i := 0; i < len(results.GraphQLResults) || i < len(results.RESTResults); i++ {
		if i < len(results.GraphQLResults) {
			r := results.GraphQLResults[i]
			status := "✅ Success"
			if !r.Success {
				status = "❌ " + r.Error
			}
			fmt.Printf("GraphQL #%-6d %-15s %-15s %-15s %s\n",
				i+1, r.CreateTime.Round(time.Millisecond),
				r.TerminateTime.Round(time.Millisecond),
				r.TotalTime.Round(time.Millisecond), status)
		}

		if i < len(results.RESTResults) {
			r := results.RESTResults[i]
			status := "✅ Success"
			if !r.Success {
				status = "❌ " + r.Error
			}
			fmt.Printf("REST #%-9d %-15s %-15s %-15s %s\n",
				i+1, r.CreateTime.Round(time.Millisecond),
				r.TerminateTime.Round(time.Millisecond),
				r.TotalTime.Round(time.Millisecond), status)
		}

		if i < len(results.GraphQLResults) && i < len(results.RESTResults) {
			fmt.Println(strings.Repeat("-", 35))
		}
	}

	// Average comparison
	fmt.Println("\n🏆 AVERAGE LATENCY COMPARISON:")
	fmt.Println(strings.Repeat("-", 70))

	if results.GraphQLAvg.Success && results.RESTAvg.Success {
		fmt.Printf("%-15s %-15s %-15s %-15s\n", "API", "Create", "Terminate", "Total")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("%-15s %-15s %-15s %-15s\n",
			"GraphQL",
			results.GraphQLAvg.CreateTime.Round(time.Millisecond),
			results.GraphQLAvg.TerminateTime.Round(time.Millisecond),
			results.GraphQLAvg.TotalTime.Round(time.Millisecond))
		fmt.Printf("%-15s %-15s %-15s %-15s\n",
			"REST",
			results.RESTAvg.CreateTime.Round(time.Millisecond),
			results.RESTAvg.TerminateTime.Round(time.Millisecond),
			results.RESTAvg.TotalTime.Round(time.Millisecond))

		// Calculate differences
		createDiff := results.RESTAvg.CreateTime - results.GraphQLAvg.CreateTime
		terminateDiff := results.RESTAvg.TerminateTime - results.GraphQLAvg.TerminateTime
		totalDiff := results.RESTAvg.TotalTime - results.GraphQLAvg.TotalTime

		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("%-15s %-15s %-15s %-15s\n",
			"Difference",
			formatDifference(createDiff),
			formatDifference(terminateDiff),
			formatDifference(totalDiff))

		// Winner analysis
		fmt.Println("\n🎯 PERFORMANCE ANALYSIS:")
		fmt.Println(strings.Repeat("-", 40))

		if totalDiff < 0 {
			fmt.Printf("🥇 Winner: REST API (%.0fms faster overall)\n", float64(-totalDiff)/float64(time.Millisecond))
		} else if totalDiff > 0 {
			fmt.Printf("🥇 Winner: GraphQL API (%.0fms faster overall)\n", float64(totalDiff)/float64(time.Millisecond))
		} else {
			fmt.Println("🤝 Tie: Both APIs have identical performance")
		}

		createPercent := float64(createDiff) / float64(results.GraphQLAvg.CreateTime) * 100
		totalPercent := float64(totalDiff) / float64(results.GraphQLAvg.TotalTime) * 100

		fmt.Printf("📈 REST is %.1f%% %s than GraphQL for pod creation\n",
			abs(createPercent),
			ternary(createPercent > 0, "slower", "faster"))
		fmt.Printf("📊 REST is %.1f%% %s than GraphQL overall\n",
			abs(totalPercent),
			ternary(totalPercent > 0, "slower", "faster"))
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
}

func formatDifference(diff time.Duration) string {
	if diff < 0 {
		return fmt.Sprintf("-%s", (-diff).Round(time.Millisecond))
	}
	return fmt.Sprintf("+%s", diff.Round(time.Millisecond))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ternary(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("RUNPOD_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set RUNPOD_API_KEY environment variable")
	}

	fmt.Println("🧪 Starting RunPod API Latency Comparison Test")
	fmt.Printf("📊 Running %d iterations for each API\n", NumIterations)
	fmt.Printf("🖼️  Test Image: %s\n", TestImage)
	fmt.Printf("🖥️  GPU Type: %s\n", TestGPUType)
	fmt.Println(strings.Repeat("-", 50))

	graphqlClient := NewGraphQLClient(apiKey)
	restClient := NewRESTClient(apiKey)

	var results ComparisonResults

	// Test GraphQL API
	fmt.Println("\n🔍 Testing GraphQL API...")
	for i := 0; i < NumIterations; i++ {
		result := testGraphQLLatency(graphqlClient, i+1)
		results.GraphQLResults = append(results.GraphQLResults, result)

		// Wait between tests to avoid rate limiting
		if i < NumIterations-1 {
			time.Sleep(3 * time.Second)
		}
	}

	// Wait between different API tests
	time.Sleep(5 * time.Second)

	// Test REST API
	fmt.Println("\n⚡ Testing REST API...")
	for i := 0; i < NumIterations; i++ {
		result := testRESTLatency(restClient, i+1)
		results.RESTResults = append(results.RESTResults, result)

		// Wait between tests to avoid rate limiting
		if i < NumIterations-1 {
			time.Sleep(3 * time.Second)
		}
	}

	// Calculate averages
	results.GraphQLAvg = calculateAverage(results.GraphQLResults)
	results.RESTAvg = calculateAverage(results.RESTResults)

	// Print results
	printResults(results)
}

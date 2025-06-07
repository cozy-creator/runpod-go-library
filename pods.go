package runpod

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CreatePod creates a new RunPod instance
func (c *Client) CreatePod(ctx context.Context, req *CreatePodRequest) (*Pod, error) {
	// Validate required fields
	if err := c.validateCreatePodRequest(req); err != nil {
		return nil, err
	}

	var pod Pod
	err := c.Post(ctx, "/pods", req, &pod)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	return &pod, nil
}

// CreateSpotPod creates a new spot/interruptible pod instance
func (c *Client) CreateSpotPod(ctx context.Context, req *CreatePodRequest, bidPerGPU float64) (*Pod, error) {
	// Set spot-specific fields
	req.BidPerGPU = bidPerGPU
	req.CloudType = "SECURE" // or "COMMUNITY" based on your needs

	return c.CreatePod(ctx, req)
}

// GetPod retrieves a pod by ID
func (c *Client) GetPod(ctx context.Context, podID string) (*Pod, error) {
	if err := c.validateRequired("podID", podID); err != nil {
		return nil, err
	}

	var pod Pod
	endpoint := fmt.Sprintf("/pods/%s", podID)
	err := c.Get(ctx, endpoint, &pod)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", podID, err)
	}

	return &pod, nil
}

// ListPods lists all pods with optional filtering
func (c *Client) ListPods(ctx context.Context, opts *ListOptions) ([]*Pod, error) {
	endpoint := c.buildListURL("/pods", opts)
	
	var response struct {
		Pods []*Pod `json:"pods"`
	}
	
	err := c.Get(ctx, endpoint, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return response.Pods, nil
}

// StopPod stops a running pod
func (c *Client) StopPod(ctx context.Context, podID string) error {
	if err := c.validateRequired("podID", podID); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/pods/%s/stop", podID)
	err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to stop pod %s: %w", podID, err)
	}

	return nil
}

// ResumePod resumes a stopped pod
func (c *Client) ResumePod(ctx context.Context, podID string) (*Pod, error) {
	if err := c.validateRequired("podID", podID); err != nil {
		return nil, err
	}

	var pod Pod
	endpoint := fmt.Sprintf("/pods/%s/resume", podID)
	err := c.Post(ctx, endpoint, nil, &pod)
	if err != nil {
		return nil, fmt.Errorf("failed to resume pod %s: %w", podID, err)
	}

	return &pod, nil
}

// TerminatePod terminates/deletes a pod
func (c *Client) TerminatePod(ctx context.Context, podID string) error {
	if err := c.validateRequired("podID", podID); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/pods/%s", podID)
	err := c.Delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to terminate pod %s: %w", podID, err)
	}

	return nil
}

// GetPodLogs retrieves logs for a pod
func (c *Client) GetPodLogs(ctx context.Context, podID string) (string, error) {
	if err := c.validateRequired("podID", podID); err != nil {
		return "", err
	}

	var response struct {
		Logs string `json:"logs"`
	}
	
	endpoint := fmt.Sprintf("/pods/%s/logs", podID)
	err := c.Get(ctx, endpoint, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", podID, err)
	}

	return response.Logs, nil
}


// GetPodStatus gets just the status of a pod (lighter than GetPod)
func (c *Client) GetPodStatus(ctx context.Context, podID string) (string, error) {
	pod, err := c.GetPod(ctx, podID)
	if err != nil {
		return "", err
	}
	return pod.Status, nil
}

// WaitForPodStatus waits for a pod to reach a specific status
func (c *Client) WaitForPodStatus(ctx context.Context, podID string, targetStatus string, maxAttempts int) (*Pod, error) {
	if maxAttempts <= 0 {
		maxAttempts = 30 // Default max attempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		pod, err := c.GetPod(ctx, podID)
		if err != nil {
			return nil, err
		}

		if strings.ToUpper(pod.Status) == strings.ToUpper(targetStatus) {
			return pod, nil
		}

		// Check if pod is in a terminal error state
		if c.isPodInErrorState(pod.Status) {
			return pod, fmt.Errorf("pod %s is in error state: %s", podID, pod.Status)
		}

		// Wait before next check
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue to next attempt
		}
	}

	return nil, fmt.Errorf("pod %s did not reach status %s after %d attempts", podID, targetStatus, maxAttempts)
}

// ListPodsByStatus lists pods filtered by status
func (c *Client) ListPodsByStatus(ctx context.Context, status string, opts *ListOptions) ([]*Pod, error) {
	pods, err := c.ListPods(ctx, opts)
	if err != nil {
		return nil, err
	}

	var filteredPods []*Pod
	for _, pod := range pods {
		if strings.ToUpper(pod.Status) == strings.ToUpper(status) {
			filteredPods = append(filteredPods, pod)
		}
	}

	return filteredPods, nil
}

// ListRunningPods lists all currently running pods
func (c *Client) ListRunningPods(ctx context.Context, opts *ListOptions) ([]*Pod, error) {
	return c.ListPodsByStatus(ctx, "RUNNING", opts)
}

// ListStoppedPods lists all stopped pods
func (c *Client) ListStoppedPods(ctx context.Context, opts *ListOptions) ([]*Pod, error) {
	return c.ListPodsByStatus(ctx, "STOPPED", opts)
}

// FindPodByName finds a pod by its name
func (c *Client) FindPodByName(ctx context.Context, name string) (*Pod, error) {
	pods, err := c.ListPods(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if pod.Name == name {
			return pod, nil
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("pod with name '%s' not found", name),
	}
}

// validateCreatePodRequest validates a pod creation request
func (c *Client) validateCreatePodRequest(req *CreatePodRequest) error {
	if req == nil {
		return NewValidationError("request", "cannot be nil")
	}

	// Required fields
	if err := c.validateRequired("name", req.Name); err != nil {
		return err
	}
	if err := c.validateRequired("imageName", req.ImageName); err != nil {
		return err
	}
	if err := c.validateRequired("gpuTypeId", req.GPUTypeID); err != nil {
		return err
	}

	// Validate positive values
	if err := c.validatePositive("gpuCount", req.GPUCount); err != nil {
		return err
	}
	if err := c.validatePositive("containerDiskInGb", req.ContainerDiskInGB); err != nil {
		return err
	}

	// Optional positive values
	if req.VCPUCount > 0 {
		if err := c.validatePositive("vcpuCount", req.VCPUCount); err != nil {
			return err
		}
	}
	if req.MemoryInGB > 0 {
		if err := c.validatePositive("memoryInGb", req.MemoryInGB); err != nil {
			return err
		}
	}
	if req.VolumeInGB > 0 {
		if err := c.validatePositive("volumeInGb", req.VolumeInGB); err != nil {
			return err
		}
	}

	// Validate bid price for spot instances
	if req.BidPerGPU > 0 {
		if err := c.validatePositiveFloat("bidPerGpu", req.BidPerGPU); err != nil {
			return err
		}
	}

	// Validate cloud type
	if req.CloudType != "" {
		validCloudTypes := []string{"SECURE", "COMMUNITY"}
		isValid := false
		for _, validType := range validCloudTypes {
			if req.CloudType == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			return NewValidationErrorWithValue("cloudType", "must be either 'SECURE' or 'COMMUNITY'", req.CloudType)
		}
	}

	return nil
}

// isPodInErrorState checks if a pod is in a terminal error state
func (c *Client) isPodInErrorState(status string) bool {
	errorStates := []string{"EXITED", "DEAD", "TERMINATED", "FAILED"}
	upperStatus := strings.ToUpper(status)
	
	for _, errorState := range errorStates {
		if upperStatus == errorState {
			return true
		}
	}
	
	return false
}


// ================================
// CONVENIENCE FUNCTIONS (Based on our scheduler usage)
// ================================

// LaunchRunPod is a convenience function that matches our existing GraphQL function signature
// This makes it easier to replace our existing code (I'll remove this function later)
func (c *Client) LaunchRunPod(ctx context.Context, imageURL string, envVars map[string]string) (string, error) {
	// Create a basic pod request with commonly used defaults
	req := &CreatePodRequest{
		Name:              fmt.Sprintf("pod-%d", time.Now().Unix()),
		ImageName:         imageURL,
		GPUTypeID:         "NVIDIA H100 80GB HBM3",
		GPUCount:          1,
		ContainerDiskInGB: 50, 
		VCPUCount:         2,  
		MemoryInGB:        15, 
		CloudType:         "SECURE",
		Env:               envVars,
		DockerArgs:        "--shm-size=1g", 
		Ports:             []string{"8080/http"}, 
		VolumeMountPath:   "/workspace", 
	}

	pod, err := c.CreatePod(ctx, req)
	if err != nil {
		return "", err
	}

	return pod.ID, nil
}

// LaunchRunPodWithConfig launches a pod with more specific configuration
func (c *Client) LaunchRunPodWithConfig(ctx context.Context, config *CreatePodRequest) (string, error) {
	pod, err := c.CreatePod(ctx, config)
	if err != nil {
		return "", err
	}

	return pod.ID, nil
}

// GetPodStatusString returns the pod status as a string (matches your existing function)
func (c *Client) GetPodStatusString(ctx context.Context, podID string) (string, error) {
	return c.GetPodStatus(ctx, podID)
}
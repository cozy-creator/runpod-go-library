package runpod

import (
	"context"
	"fmt"
	"time"
	"reflect"
)


// =========================
// SERVERLESS JOB OPERATIONS
// =========================

// RunAsync submits an asynchronous job to a serverless endpoint
// Returns immediately with a job ID for later status checking
func (c *Client) RunAsync(ctx context.Context, endpointID string, input interface{}) (*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}

	req := &RunJobRequest{Input: input}

	endpoint := fmt.Sprintf("/v2/%s/run", endpointID)

	var job Job
	err := c.Post(ctx, endpoint, req, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to submit async job to endpoint %s: %w", endpointID, err)
	}

	return &job, nil
}

// RunSync submits a synchronous job and waits for completion
// Blocks until the job completes or times out
func (c *Client) RunSync(ctx context.Context, endpointID string, input interface{}) (*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}

	req := &RunJobRequest{Input: input}
	
	endpoint := fmt.Sprintf("/v2/%s/runsync", endpointID)
	
	var job Job
	err := c.Post(ctx, endpoint, req, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to submit sync job to endpoint %s: %w", endpointID, err)
	}

	return &job, nil
}

// GetJobStatus retrieves the status and results of a job
func (c *Client) GetJobStatus(ctx context.Context, endpointID, jobID string) (*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}
	if err := c.validateRequired("jobID", jobID); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/v2/%s/status/%s", endpointID, jobID)
	
	var job Job
	err := c.Get(ctx, endpoint, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to get status for job %s on endpoint %s: %w", jobID, endpointID, err)
	}

	return &job, nil
}

// CancelJob cancels a running or queued job
func (c *Client) CancelJob(ctx context.Context, endpointID, jobID string) error {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return err
	}
	if err := c.validateRequired("jobID", jobID); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/v2/%s/cancel/%s", endpointID, jobID)
	
	err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel job %s on endpoint %s: %w", jobID, endpointID, err)
	}

	return nil
}

// RetryJob retries a failed or timed-out job using the same job ID and input
func (c *Client) RetryJob(ctx context.Context, endpointID, jobID string) (*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}
	if err := c.validateRequired("jobID", jobID); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/v2/%s/retry/%s", endpointID, jobID)
	
	var job Job
	err := c.Post(ctx, endpoint, nil, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to retry job %s on endpoint %s: %w", jobID, endpointID, err)
	}

	return &job, nil
}

// PurgeQueue clears all pending jobs from the endpoint queue
// Does not affect jobs that are already running
func (c *Client) PurgeQueue(ctx context.Context, endpointID string) error {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/v2/%s/purge-queue", endpointID)
	
	err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to purge queue for endpoint %s: %w", endpointID, err)
	}

	return nil
}

// GetHealth checks the operational status of an endpoint
func (c *Client) GetHealth(ctx context.Context, endpointID string) (*EndpointHealth, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/v2/%s/health", endpointID)
	
	var health EndpointHealth
	err := c.Get(ctx, endpoint, &health)
	if err != nil {
		return nil, fmt.Errorf("failed to get health for endpoint %s: %w", endpointID, err)
	}

	return &health, nil
}

// ================================
// JOB MONITORING AND UTILITIES
// ================================

// WaitForJobCompletion waits for a job to complete or fail
// Returns the final job state or an error if timeout is reached
func (c *Client) WaitForJobCompletion(ctx context.Context, endpointID, jobID string, maxWaitTime time.Duration) (*Job, error) {
	if maxWaitTime <= 0 {
		maxWaitTime = 10 * time.Minute // Default timeout
	}

	deadline := time.Now().Add(maxWaitTime)
	
	for time.Now().Before(deadline) {
		job, err := c.GetJobStatus(ctx, endpointID, jobID)
		if err != nil {
			return nil, err
		}

		// Check if job is in a terminal state
		switch JobStatus(job.Status) {
		case JobStatusCompleted:
			return job, nil
		case JobStatusFailed:
			return job, fmt.Errorf("job %s failed: %s", jobID, job.Error)
		case JobStatusCancelled:
			return job, fmt.Errorf("job %s was cancelled", jobID)
		case JobStatusTimedOut:
			return job, fmt.Errorf("job %s timed out", jobID)
		}

		// Wait before next check
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue polling
		}
	}

	return nil, fmt.Errorf("job %s did not complete within %v", jobID, maxWaitTime)
}

// IsJobTerminal checks if a job is in a terminal state (completed, failed, etc.)
func (c *Client) IsJobTerminal(status string) bool {
	terminalStates := []JobStatus{
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusCancelled,
		JobStatusTimedOut,
	}

	jobStatus := JobStatus(status)
	for _, terminalStatus := range terminalStates {
		if jobStatus == terminalStatus {
			return true
		}
	}

	return false
}

// ================================
// BATCH JOB OPERATIONS
// ================================

// SubmitMultipleJobs submits multiple jobs to the same endpoint asynchronously
func (c *Client) SubmitMultipleJobs(ctx context.Context, endpointID string, inputs []interface{}) ([]*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, NewValidationError("inputs", "cannot be empty")
	}

	var jobs []*Job
	var errors []error

	for i, input := range inputs {
		job, err := c.RunAsync(ctx, endpointID, input)
		if err != nil {
			errors = append(errors, fmt.Errorf("job %d failed: %w", i, err))
			continue
		}
		jobs = append(jobs, job)
	}

	// If any jobs failed to submit, return error with details
	if len(errors) > 0 {
		return jobs, fmt.Errorf("failed to submit %d out of %d jobs: %v", len(errors), len(inputs), errors)
	}

	return jobs, nil
}

// WaitForMultipleJobs waits for multiple jobs to complete
func (c *Client) WaitForMultipleJobs(ctx context.Context, endpointID string, jobIDs []string, maxWaitTime time.Duration) ([]*Job, error) {
	if len(jobIDs) == 0 {
		return nil, NewValidationError("jobIDs", "cannot be empty")
	}

	results := make([]*Job, len(jobIDs))
	completed := make([]bool, len(jobIDs))
	
	deadline := time.Now().Add(maxWaitTime)

	for time.Now().Before(deadline) {
		allCompleted := true

		for i, jobID := range jobIDs {
			if completed[i] {
				continue // Already completed
			}

			job, err := c.GetJobStatus(ctx, endpointID, jobID)
			if err != nil {
				return results, fmt.Errorf("failed to get status for job %s: %w", jobID, err)
			}

			if c.IsJobTerminal(job.Status) {
				results[i] = job
				completed[i] = true
			} else {
				allCompleted = false
			}
		}

		if allCompleted {
			return results, nil
		}

		// Wait before next check
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue polling
		}
	}

	// Count how many completed
	completedCount := 0
	for _, isCompleted := range completed {
		if isCompleted {
			completedCount++
		}
	}

	return results, fmt.Errorf("%d out of %d jobs completed within %v", completedCount, len(jobIDs), maxWaitTime)
}

// ================================
// STREAMING SUPPORT
// ================================

// StreamResults retrieves partial/streaming results from a job
// This is useful for jobs that generate output incrementally (like text generation)
// or have very large outputs that benefit from chunked delivery
func (c *Client) StreamResults(ctx context.Context, endpointID, jobID string) (*Job, error) {
	if err := c.validateRequired("endpointID", endpointID); err != nil {
		return nil, err
	}
	if err := c.validateRequired("jobID", jobID); err != nil {
		return nil, err
	}

	// RunPod stream endpoint: /v2/{endpoint_id}/stream/{job_id}
	endpoint := fmt.Sprintf("/v2/%s/stream/%s", endpointID, jobID)
	
	var job Job
	err := c.Get(ctx, endpoint, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to stream results for job %s on endpoint %s: %w", jobID, endpointID, err)
	}

	return &job, nil
}

// StreamResultsContinuous polls the stream endpoint for continuous updates
// Returns channels for job updates and errors - useful for long-running jobs
// This provides a convenient wrapper around StreamResults for real-time monitoring
func (c *Client) StreamResultsContinuous(ctx context.Context, endpointID, jobID string, pollInterval time.Duration) (<-chan *Job, <-chan error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second // Default poll interval
	}

	jobChan := make(chan *Job, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(jobChan)
		defer close(errChan)

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		var lastOutput interface{}
		
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case <-ticker.C:
				job, err := c.StreamResults(ctx, endpointID, jobID)
				if err != nil {
					errChan <- err
					return
				}

				// Send update if output has changed or status changed
				outputChanged := !compareOutputs(lastOutput, job.Output)
				if outputChanged {
					select {
					case jobChan <- job:
						lastOutput = job.Output
					case <-ctx.Done():
						return
					}
				}

				// Stop streaming if job is terminal
				if c.IsJobTerminal(job.Status) {
					return
				}
			}
		}
	}()

	return jobChan, errChan
}

// compareOutputs compares two job outputs to detect changes
func compareOutputs(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// ================================
// CONVENIENCE FUNCTIONS
// ================================

// RunAndWait submits a job asynchronously and waits for completion
// Combines RunAsync + WaitForJobCompletion for convenience
func (c *Client) RunAndWait(ctx context.Context, endpointID string, input interface{}, maxWaitTime time.Duration) (*Job, error) {
	// Submit job
	job, err := c.RunAsync(ctx, endpointID, input)
	if err != nil {
		return nil, err
	}

	// Wait for completion
	return c.WaitForJobCompletion(ctx, endpointID, job.ID, maxWaitTime)
}

// QuickRun is a convenience function for simple synchronous job execution
// Uses reasonable defaults for timeout and error handling
func (c *Client) QuickRun(ctx context.Context, endpointID string, input interface{}) (*Job, error) {
	// Try sync first (faster for quick jobs)
	job, err := c.RunSync(ctx, endpointID, input)
	if err != nil {
		// If sync fails, try async with wait
		return c.RunAndWait(ctx, endpointID, input, 5*time.Minute)
	}
	return job, nil
}




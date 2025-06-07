package runpod

import "time"

type ListOptions struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type Pod struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	ImageName         string            `json:"imageName"`
	GPUCount          int               `json:"gpuCount"`
	GPUTypeID         string            `json:"gpuTypeId"`
	VCPUCount         int               `json:"vcpuCount"`
	MemoryInGB        int               `json:"memoryInGb"`
	ContainerDiskInGB int               `json:"containerDiskInGb"`
	VolumeInGB        int               `json:"volumeInGb"`
	VolumeMountPath   string            `json:"volumeMountPath"`
	CostPerHour       float64           `json:"costPerHour"`
	DatacenterID      string            `json:"datacenterId"`
	CreatedAt         time.Time         `json:"createdAt"`
	Env               map[string]string `json:"env"`
	Ports             []string          `json:"ports"`
	DockerArgs        string            `json:"dockerArgs"`
	Runtime           *PodRuntime       `json:"runtime"`
	MachineID         string            `json:"machineId"`
	LastStartAt       *time.Time        `json:"lastStartAt"`
}

type PodRuntime struct {
	UptimeSeconds    int    `json:"uptimeSeconds"`
	LastStartedAt    string `json:"lastStartedAt"`
	LastStatusCharge string `json:"lastStatusCharge"`
}

type CreatePodRequest struct {
	Name              string            `json:"name"`
	ImageName         string            `json:"imageName"`
	GPUTypeID         string            `json:"gpuTypeId"`
	GPUCount          int               `json:"gpuCount"`
	VCPUCount         int               `json:"vcpuCount,omitempty"`
	MemoryInGB        int               `json:"memoryInGb,omitempty"`
	ContainerDiskInGB int               `json:"containerDiskInGb"`
	VolumeInGB        int               `json:"volumeInGb,omitempty"`
	VolumeMountPath   string            `json:"volumeMountPath,omitempty"`
	DatacenterID      string            `json:"datacenterId,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
	Ports             []string          `json:"ports,omitempty"`
	DockerArgs        string            `json:"dockerArgs,omitempty"`
	BidPerGPU         float64           `json:"bidPerGpu,omitempty"` // For spot instances
	CloudType         string            `json:"cloudType,omitempty"` // "SECURE" or "COMMUNITY"
	NetworkVolumeID   string            `json:"networkVolumeId,omitempty"`
}

type UpdatePodRequest struct {
	Name       string            `json:"name,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	DockerArgs string            `json:"dockerArgs,omitempty"`
	BidPerGPU  float64           `json:"bidPerGpu,omitempty"` // For spot instances
}

type Endpoint struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	TemplateID       string    `json:"templateId"`
	GPUTypeIDs       []string  `json:"gpuTypeIds"`
	ScalerType       string    `json:"scalerType"`
	ScalerValue      int       `json:"scalerValue"`
	WorkersMin       int       `json:"workersMin"`
	WorkersMax       int       `json:"workersMax"`
	IdleTimeout      int       `json:"idleTimeout"`
	ExecutionTimeout int       `json:"executionTimeoutMs"`
	CreatedAt        time.Time `json:"createdAt"`
	Status           string    `json:"status"`
	URL              string    `json:"url,omitempty"`
}

type CreateEndpointRequest struct {
	Name             string   `json:"name"`
	TemplateID       string   `json:"templateId"`
	GPUTypeIDs       []string `json:"gpuTypeIds"`
	ScalerType       string   `json:"scalerType"`
	ScalerValue      int      `json:"scalerValue"`
	WorkersMin       int      `json:"workersMin"`
	WorkersMax       int      `json:"workersMax"`
	IdleTimeout      int      `json:"idleTimeout"`
	ExecutionTimeout int      `json:"executionTimeoutMs"`
}

type UpdateEndpointRequest struct {
	Name             string   `json:"name,omitempty"`
	GPUTypeIDs       []string `json:"gpuTypeIds,omitempty"`
	ScalerType       string   `json:"scalerType,omitempty"`
	ScalerValue      int      `json:"scalerValue,omitempty"`
	WorkersMin       int      `json:"workersMin,omitempty"`
	WorkersMax       int      `json:"workersMax,omitempty"`
	IdleTimeout      int      `json:"idleTimeout,omitempty"`
	ExecutionTimeout int      `json:"executionTimeoutMs,omitempty"`
}

type Job struct {
	ID             string      `json:"id"`
	Status         string      `json:"status"`
	Input          interface{} `json:"input"`
	Output         interface{} `json:"output,omitempty"`
	Error          string      `json:"error,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	StartedAt      *time.Time  `json:"startedAt,omitempty"`
	CompletedAt    *time.Time  `json:"completedAt,omitempty"`
	ExecutionTime  int         `json:"executionTimeMs,omitempty"`
	RetryCount     int         `json:"retryCount,omitempty"`
	EndpointID     string      `json:"endpointId,omitempty"`
}

type RunJobRequest struct {
	Input interface{} `json:"input"`
}

type JobStatus string

const (
	JobStatusInQueue    JobStatus = "IN_QUEUE"
	JobStatusInProgress JobStatus = "IN_PROGRESS" 
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
	JobStatusTimedOut   JobStatus = "TIMED_OUT"
)

type Template struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	ImageName         string            `json:"imageName"`
	IsServerless      bool              `json:"isServerless"`
	ContainerDiskInGB int               `json:"containerDiskInGb"`
	VolumeInGB        int               `json:"volumeInGb"`
	VolumeMountPath   string            `json:"volumeMountPath"`
	Env               map[string]string `json:"env"`
	Ports             string            `json:"ports"`
	DockerArgs        string            `json:"dockerArgs"`
	CreatedAt         time.Time         `json:"createdAt"`
	Runtime           *TemplateRuntime  `json:"runtime,omitempty"`
}

type TemplateRuntime struct {
	ContainerRegistryAuthID string `json:"containerRegistryAuthId,omitempty"`
	StartSSH                bool   `json:"startSsh,omitempty"`
}

type CreateTemplateRequest struct {
	Name              string            `json:"name"`
	ImageName         string            `json:"imageName"`
	IsServerless      bool              `json:"isServerless"`
	ContainerDiskInGB int               `json:"containerDiskInGb"`
	VolumeInGB        int               `json:"volumeInGb,omitempty"`
	VolumeMountPath   string            `json:"volumeMountPath,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
	Ports             string            `json:"ports,omitempty"`
	DockerArgs        string            `json:"dockerArgs,omitempty"`
	Runtime           *TemplateRuntime  `json:"runtime,omitempty"`
}

type UpdateTemplateRequest struct {
	Name              string            `json:"name,omitempty"`
	ImageName         string            `json:"imageName,omitempty"`
	ContainerDiskInGB int               `json:"containerDiskInGb,omitempty"`
	VolumeInGB        int               `json:"volumeInGb,omitempty"`
	VolumeMountPath   string            `json:"volumeMountPath,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
	Ports             string            `json:"ports,omitempty"`
	DockerArgs        string            `json:"dockerArgs,omitempty"`
	Runtime           *TemplateRuntime  `json:"runtime,omitempty"`
}

type GPUType struct {
	ID             string  `json:"id"`
	DisplayName    string  `json:"displayName"`
	MemoryInGB     int     `json:"memoryInGb"`
	CostPerHour    float64 `json:"costPerHr"`
	Available      bool    `json:"available"`
	CommunityCloud bool    `json:"communityCloud"`
	SecureCloud    bool    `json:"secureCloud"`
	LowestPrice    *Price  `json:"lowestPrice,omitempty"`
}

type Price struct {
	MinimumBidPrice float64 `json:"minimumBidPrice"`
	UninterruptablePrice float64 `json:"uninterruptablePrice"`
	InterruptablePrice   float64 `json:"interruptablePrice,omitempty"`
}

type Datacenter struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Region  string `json:"region,omitempty"`
}

type AccountInfo struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	Balance          float64 `json:"balance"`
	SpendLimit       float64 `json:"spendLimit,omitempty"`
	CurrentSpendPerHr float64 `json:"currentSpendPerHr"`
	MachineQuota     int     `json:"machineQuota,omitempty"`
}

// NetworkVolume represents a network volume
type NetworkVolume struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Size         int       `json:"size"`
	DatacenterID string    `json:"datacenterId"`
	CreatedAt    time.Time `json:"createdAt"`
	PodIds       []string  `json:"podIds,omitempty"`
}

type CreateNetworkVolumeRequest struct {
	Name         string `json:"name"`
	Size         int    `json:"size"`
	DatacenterID string `json:"datacenterId"`
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

type EndpointHealth struct {
	Status        string `json:"status"`
	JobsInQueue   int    `json:"jobsInQueue"`
	WorkersIdle   int    `json:"workersIdle"`
	WorkersActive int    `json:"workersActive"`
	WorkersTotal  int    `json:"workersTotal"`
}

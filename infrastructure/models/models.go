package models

type TaskRequest struct {
	Image string            `json:"image"`
	Name  string            `json:"name"`
	Env   map[string]string `json:"env"`
	Ports map[string]int    `json:"ports"`
}

type TaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type NATSInvocation struct {
	TaskID    string `json:"task_id"`
	Image     string `json:"image"` // Dynamic image
	Execution struct {
		Kind    string   `json:"kind"`
		Path    string   `json:"path"`
		Command []string `json:"command"` // Generic command override
	} `json:"execution"`
	Resources struct {
		CPU    float64 `json:"cpu"`
		Memory int64   `json:"memory"`
	} `json:"resources"`
	TimeoutMS int                    `json:"timeout_ms"`
	Env       map[string]string      `json:"env"`
	Payload   map[string]interface{} `json:"payload"`
}

type NATSResponse struct {
	TaskID          string `json:"task_id"`
	Status          string `json:"status"`
	ExecutionResult string `json:"execution_result"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
}

type NATSStatusUpdate struct {
	TaskID  string        `json:"task_id"`
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Result  *NATSResponse `json:"result,omitempty"`
}

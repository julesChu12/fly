package saga

import (
	"time"
)

// SagaEventType Saga事件类型
type SagaEventType string

const (
	SagaEventStarted               SagaEventType = "saga_started"
	SagaEventCompleted             SagaEventType = "saga_completed"
	SagaEventFailed                SagaEventType = "saga_failed"
	SagaEventCompensationStarted   SagaEventType = "saga_compensation_started"
	SagaEventCompensationCompleted SagaEventType = "saga_compensation_completed"
	SagaEventCancelled             SagaEventType = "saga_cancelled"
)

// SagaStepEventType Saga步骤事件类型
type SagaStepEventType string

const (
	SagaStepStarted     SagaStepEventType = "saga_step_started"
	SagaStepCompleted   SagaStepEventType = "saga_step_completed"
	SagaStepFailed      SagaStepEventType = "saga_step_failed"
	SagaStepCompensated SagaStepEventType = "saga_step_compensated"
	SagaStepSkipped     SagaStepEventType = "saga_step_skipped"
)

// SagaEvent Saga事件
type SagaEvent struct {
	ID        string        `json:"id"`
	Type      SagaEventType `json:"type"`
	SagaID    string        `json:"saga_id"`
	SagaName  string        `json:"saga_name"`
	Status    SagaStatus    `json:"status"`
	Payload   interface{}   `json:"payload"`
	Timestamp time.Time     `json:"timestamp"`
}

// SagaStepEvent Saga步骤事件
type SagaStepEvent struct {
	ID        string            `json:"id"`
	Type      SagaStepEventType `json:"type"`
	SagaID    string            `json:"saga_id"`
	SagaName  string            `json:"saga_name"`
	StepID    string            `json:"step_id"`
	StepName  string            `json:"step_name"`
	Status    StepStatus        `json:"status"`
	Result    interface{}       `json:"result,omitempty"`
	Error     error             `json:"error,omitempty"`
	Payload   interface{}       `json:"payload,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

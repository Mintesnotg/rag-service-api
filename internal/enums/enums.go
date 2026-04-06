package enums

type DocumentStatus string

const (
	DocumentStatusActive   DocumentStatus = "active"
	DocumentStatusInactive DocumentStatus = "inactive"
)

type ProcessingStatus string

const (
	ProcessingPending   ProcessingStatus = "pending"
	ProcessingRunning   ProcessingStatus = "processing"
	ProcessingCompleted ProcessingStatus = "completed"
	ProcessingFailed    ProcessingStatus = "failed"
)

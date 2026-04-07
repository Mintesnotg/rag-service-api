package enums

// RecordStatus represents the active/inactive state of a record.
type RecordStatus string

const (
	StatusActive   RecordStatus = "active"
	StatusInactive RecordStatus = "inactive"
)

// DocumentStatus is kept as an alias for backward compatibility, if needed elsewhere.
type DocumentStatus = RecordStatus

type ProcessingStatus string

const (
	ProcessingPending   ProcessingStatus = "pending"
	ProcessingRunning   ProcessingStatus = "processing"
	ProcessingCompleted ProcessingStatus = "completed"
	ProcessingFailed    ProcessingStatus = "failed"
)

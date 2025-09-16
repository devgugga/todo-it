package enums

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusCancelled  TaskStatus = "cancelled"
)

func (s TaskStatus) String() string {
	return string(s)
}

func GetAllStatuses() []TaskStatus {
	return []TaskStatus{
		StatusPending,
		StatusInProgress,
		StatusCompleted,
		StatusCancelled,
	}
}

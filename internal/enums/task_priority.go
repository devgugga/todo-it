package enums

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
	PriorityUrgent TaskPriority = "urgent"
)

func (p TaskPriority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func (p TaskPriority) String() string {
	return string(p)
}

func GetAllPriorities() []TaskPriority {
	return []TaskPriority{
		PriorityLow,
		PriorityMedium,
		PriorityHigh,
		PriorityUrgent,
	}
}

func (p TaskPriority) GetPriorityOrder() int {
	switch p {
	case PriorityLow:
		return 1
	case PriorityMedium:
		return 2
	case PriorityHigh:
		return 3
	case PriorityUrgent:
		return 4
	default:
		return 0
	}
}

package domain

type (
	ActivityType int
)

var (
	ActivityTaskField          = ActivityType(1)
	ActivityTaskStatus         = ActivityType(2)
	ActivityTaskName           = ActivityType(3)
	ActivityTaskParent         = ActivityType(4)
	ActivityTaskFieldArray     = ActivityType(5)
	ActivityTaskTeamArray      = ActivityType(6)
	ActivityTaskWasDeleted     = ActivityType(8)
	ActivityTaskFileWasDeleted = ActivityType(9)
)

package store

type OperationType string

type TypeCommand string

const (
	Admin OperationType = "admin"
)

const (
	AdminCreate TypeCommand = "create"
	AdminDelete TypeCommand = "delete"
)

var MapTypes = map[TypeCommand]OperationType{
	AdminCreate: Admin,
	AdminDelete: Admin,
}

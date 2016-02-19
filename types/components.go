package types

type ComponentType string

const (
	ComponentService ComponentType = "Service"
	ComponentUPS                   = "User provided service"
	ComponentApp                   = "Application"
)

type Component struct {
	GUID         string
	Name         string
	Type         ComponentType
	DependencyOf []string
	Clone        bool
}

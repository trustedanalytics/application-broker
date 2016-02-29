package types

type ComponentType string

const (
	ComponentService ComponentType = "Service"
	ComponentUPS                   = "User provided service"
	ComponentApp                   = "Application"
)

type Component struct {
	GUID         string        `json:"GUID"`
	Name         string        `json:"name"`
	Type         ComponentType `json:"type"`
	DependencyOf []string      `json:"dependencyOf"`
	Clone        bool          `json:"clone"`
}

type ComponentClone struct {
	Component Component `json:"component"`
	CloneGUID string    `json:"cloneGUID"`
}

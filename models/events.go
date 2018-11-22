package models

type event string

var (
	Create = "create"
	Update = "update"
	Delete = "delete"

	AllEvents = []string{
		Create, Update, Delete,
	}

	ValidEventMap = map[string]string{
		Create: Create,
		Update: Update,
		Delete: Delete,
	}
)

type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

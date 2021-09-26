package types

type (
	dataType string

	// PreWriteMessage represents an event that asks a worker to write data to the datastore.
	PreWriteMessage struct {
		DataType                dataType                      `json:"dataType"`
		Item                    *ItemDatabaseCreationInput    `json:"item,omitempty"`
		Webhook                 *WebhookDatabaseCreationInput `json:"webhook,omitempty"`
		UserMembership          *AddUserToAccountInput        `json:"user_membership"`
		AttributableToUserID    string                        `json:"attributableToUserID"`
		AttributableToAccountID string                        `json:"attributeToAccountID"`
	}

	// PreUpdateMessage represents an event that asks a worker to update data to the datastore.
	PreUpdateMessage struct {
		DataType                dataType `json:"dataType"`
		Item                    *Item    `json:"item,omitempty"`
		AttributableToUserID    string   `json:"attributableToUserID"`
		AttributableToAccountID string   `json:"attributeToAccountID"`
	}

	// PreArchiveMessage represents an event that asks a worker to archive data to the datastore.
	PreArchiveMessage struct {
		DataType                dataType `json:"dataType"`
		RelevantID              string   `json:"relevantID"`
		AttributableToUserID    string   `json:"attributableToUserID"`
		AttributableToAccountID string   `json:"attributeToAccountID"`
	}

	// DataChangeMessage represents an event that asks a worker to write data to the datastore.
	DataChangeMessage struct {
		MessageType             string                 `json:"messageType"`
		DataType                dataType               `json:"dataType"`
		Item                    *Item                  `json:"item,omitempty"`
		Webhook                 *Webhook               `json:"webhook,omitempty"`
		UserMembership          *AccountUserMembership `json:"user_membership"`
		Context                 map[string]string      `json:"context"`
		AttributableToUserID    string                 `json:"attributableToUserID"`
		AttributableToAccountID string                 `json:"attributeToAccountID"`
	}
)

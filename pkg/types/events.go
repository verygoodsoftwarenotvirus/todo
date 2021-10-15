package types

type (
	dataType string

	// PreWriteMessage represents an event that asks a worker to write data to the datastore.
	PreWriteMessage struct {
		_ struct{}

		DataType                dataType                      `json:"dataType"`
		Item                    *ItemDatabaseCreationInput    `json:"item,omitempty"`
		Webhook                 *WebhookDatabaseCreationInput `json:"webhook,omitempty"`
		UserMembership          *AddUserToAccountInput        `json:"userMembership,omitempty"`
		AttributableToUserID    string                        `json:"attributableToUserID"`
		AttributableToAccountID string                        `json:"attributableToAccountID"`
	}

	// PreUpdateMessage represents an event that asks a worker to update data in the datastore.
	PreUpdateMessage struct {
		_ struct{}

		DataType                dataType `json:"dataType"`
		Item                    *Item    `json:"item,omitempty"`
		AttributableToUserID    string   `json:"attributableToUserID"`
		AttributableToAccountID string   `json:"attributableToAccountID"`
	}

	// PreArchiveMessage represents an event that asks a worker to archive data in the datastore.
	PreArchiveMessage struct {
		_ struct{}

		DataType                dataType `json:"dataType"`
		ItemID                  string   `json:"itemID"`
		WebhookID               string   `json:"webhookID"`
		AttributableToUserID    string   `json:"attributableToUserID"`
		AttributableToAccountID string   `json:"attributableToAccountID"`
	}

	// DataChangeMessage represents an event that asks a worker to write data to the datastore.
	DataChangeMessage struct {
		_ struct{}

		DataType                dataType               `json:"dataType"`
		MessageType             string                 `json:"messageType"`
		Item                    *Item                  `json:"item,omitempty"`
		Webhook                 *Webhook               `json:"webhook,omitempty"`
		UserMembership          *AccountUserMembership `json:"userMembership,omitempty"`
		Context                 map[string]string      `json:"context,omitempty"`
		AttributableToUserID    string                 `json:"attributableToUserID"`
		AttributableToAccountID string                 `json:"attributableToAccountID"`
	}
)

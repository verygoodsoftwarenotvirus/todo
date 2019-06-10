package models

// DataExport represents a user's data export
type DataExport struct {
	User User
	Items []Item
	Webhooks []Webhook
	OAuth2Clients []OAuth2Client
}

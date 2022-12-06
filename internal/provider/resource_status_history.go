package provider

type statusHistory struct {
	Day              *string `json:"day,omitempty"`
	Status           *string `json:"status,omitempty"`
	DowntimeDuration *int    `json:"downtime_duration,omitempty"`
}

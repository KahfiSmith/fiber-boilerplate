package response

type HealthStatusResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

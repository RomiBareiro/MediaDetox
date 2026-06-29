package httpadapter

type registerEventRequest struct {
	UserID    string `json:"user_id"`
	AppName   string `json:"app_name"`
	EventType string `json:"event_type"`
	Timestamp string `json:"timestamp"`
}

type createRuleRequest struct {
	AppName       string `json:"app_name"`
	MaxOpens      int    `json:"max_opens"`
	WindowMinutes int    `json:"window_minutes"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

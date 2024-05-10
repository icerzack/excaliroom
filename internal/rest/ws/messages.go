package ws

type JWTValidationResponse struct {
	ID int `json:"id"`
}

type Message struct {
	Event string `json:"event"`
}

type MessageConnectRequest struct {
	Message
	BoardID string `json:"board_id"`
	Jwt     string `json:"jwt"`
}

type MessageNewDataRequest struct {
	Message
	BoardID string `json:"board_id"`
	Jwt     string `json:"jwt"`
	Data    Data   `json:"data"`
}

type MessageUserConnectedResponse struct {
	Message
	BoardID  string   `json:"board_id"`
	UserIDs  []string `json:"user_ids"`
	LeaderID string   `json:"leader_id"`
}

type MessageSetLeaderRequest struct {
	Message
	BoardID string `json:"board_id"`
	Jwt     string `json:"jwt"`
}

type MessageSetLeaderResponse struct {
	Message
	BoardID string `json:"board_id"`
	UserID  string `json:"user_id"`
}

type MessageUserFailedToConnectResponse struct {
	Message
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type MessageUserDisconnectedResponse struct {
	Message
	BoardID  string   `json:"board_id"`
	UserIDs  []string `json:"user_ids"`
	LeaderID string   `json:"leader_id"`
}

type MessageNewDataResponse struct {
	Message
	BoardID string `json:"board_id"`
	Data    Data   `json:"data"`
}

type Data struct {
	Elements string `json:"elements"`
	AppState string `json:"app_state"`
}

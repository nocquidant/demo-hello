package api

type MsgResponse struct {
	Msg string `json:"msg"`
}

type HealthResponse struct {
	Health string `json:"health"`
}

type MsgRemoteResponse struct {
	Msg        string      `json:"msg"`
	FromRemote MsgResponse `json:"fromRemote"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

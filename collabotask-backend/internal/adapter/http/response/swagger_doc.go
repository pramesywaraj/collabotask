package response

type DocSuccessEnvelope struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type AuthSuccessDoc struct {
	DocSuccessEnvelope
	Data AuthResponse `json:"data"`
}

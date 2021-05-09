package model

type ErrorResponse struct {
	Error_Message string `json:"error_message,omitempty"`
	Error_Code    int    `json:"error_code,omitempty"`
}

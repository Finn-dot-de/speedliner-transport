package structs

// MailRecipient für Swagger mit Enums/Examples
type MailRecipient struct {
	ID   int64  `json:"id"   example:"1234546456"`
	Type string `json:"type" enums:"character,corporation,alliance,mailing_list" example:"character"`
}

// SendMailRequest für Swagger (AutoApproveCspa bleibt drin, default=false)
type SendMailRequest struct {
	Subject         string          `json:"subject"               example:"Test"`
	Body            string          `json:"body"                  example:"Hello World"`
	Recipients      []MailRecipient `json:"recipients"`
	AutoApproveCspa bool            `json:"autoApproveCspa"       example:"false"`
}

type MailIDResponse struct {
	MailID int `json:"mail_id" example:"399492427"`
}
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid JSON"`
}

// @Success 201 {object} handler.MailIDResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Failure 502 {object} handler.ErrorResponse

package private

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.gitea.io/gitea/modules/setting"
)

// Email structure holds a data for sending general emails
type Email struct {
	Subject string
	Message string
	To      []string
}

// SendEmail calls the internal SendEmail function
//
// It accepts a list of usernames.
// If DB contains these users it will send the email to them.
//
// If to list == nil its supposed to send an email to every
// user present in DB
func SendEmail(subject, message string, to []string) (int, string) {
	reqURL := setting.LocalURL + "api/internal/mail/send"

	req := newInternalRequest(reqURL, "POST")
	req = req.Header("Content-Type", "application/json")
	jsonBytes, _ := json.Marshal(Email{
		Subject: subject,
		Message: message,
		To:      to,
	})
	req.Body(jsonBytes)
	resp, err := req.Response()
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Unable to contact gitea: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Responce body error: %v", err.Error())
	}

	return http.StatusOK, fmt.Sprintf("Was sent %s from %d", body, len(to))
}

package private

import (
	"fmt"
	"net/http"
	"strconv"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/private"
	"code.gitea.io/gitea/services/mailer"
	"gitea.com/macaron/macaron"
)

// SendEmail pushes messages to mail queue
//
// It doesn't wait before each message will be processed
func SendEmail(ctx *macaron.Context, mail private.Email) {
	var emails []string
	if len(mail.To) > 0 {
		for _, uname := range mail.To {
			user, err := models.GetUserByName(uname)
			if err != nil {
				err := fmt.Sprintf("Failed to get user information: %v", err)
				log.Error(err)
				ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"err": err,
				})
				return
			}

			if user != nil {
				emails = append(emails, user.Email)
			}
		}
	} else {
		err := models.IterateUser(func(user *models.User) error {
			emails = append(emails, user.Email)
			return nil
		})
		if err != nil {
			err := fmt.Sprintf("Failed to find users: %v", err)
			log.Error(err)
			ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"err": err,
			})
			return
		}
	}

	sendEmail(ctx, mail.Subject, mail.Message, emails)
}

func sendEmail(ctx *macaron.Context, subject, message string, to []string) {
	for _, email := range to {
		msg := mailer.NewMessage([]string{email}, subject, message)
		mailer.SendAsync(msg)
	}

	wasSent := strconv.Itoa(len(to))

	ctx.PlainText(http.StatusOK, []byte(wasSent))
}

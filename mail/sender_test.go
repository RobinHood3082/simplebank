package mail

import (
	"testing"

	"github.com/RobinHood3082/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := "This is a test message from <a href=\"https://github.com/RobinHood3082/simplebank\">Simple Bank</a>"
	to := []string{"random_email@gmail.com"}
	cc := []string{}
	bcc := []string{}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, cc, bcc, attachFiles)
	require.NoError(t, err)
}

package services

import (
	"fmt"
	"net/smtp"
	// "github.com/resend/resend-go/v4"
)

func buildOTPEmailHTML(otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Report It Verification Code</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f5f7; font-family:'Segoe UI', Roboto, Helvetica, Arial, sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f5f7; padding:40px 0;">
    <tr>
      <td align="center">
        <table role="presentation" width="480" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border:1px solid #e4e6eb; border-radius:8px;">
 
          <!-- Header -->
          <tr>
            <td style="padding:28px 32px; border-bottom:1px solid #e4e6eb;">
              <span style="color:#1a1a1a; font-size:18px; font-weight:600;">Report It</span>
            </td>
          </tr>
 
          <!-- Body -->
          <tr>
            <td style="padding:32px;">
              <h1 style="margin:0 0 12px 0; color:#1a1a1a; font-size:20px; font-weight:600;">Verify your email</h1>
              <p style="margin:0 0 28px 0; color:#5f6368; font-size:14px; line-height:1.6;">
                Use the verification code below to continue with your Report It account. This code is valid for 5 minutes.
              </p>
 
              <!-- OTP block -->
              <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f5f7; border-radius:6px;">
                <tr>
                  <td style="padding:22px; text-align:center;">
                    <span style="font-family:'Courier New', Courier, monospace; font-size:32px; font-weight:700; letter-spacing:8px; color:#1a1a1a;">%s</span>
                  </td>
                </tr>
              </table>
 
              <p style="margin:24px 0 0 0; color:#5f6368; font-size:13px; line-height:1.6;">
                If you didn't request this code, you can safely ignore this email.
              </p>
            </td>
          </tr>
 
          <!-- Footer -->
          <tr>
            <td style="padding:20px 32px; border-top:1px solid #e4e6eb; text-align:center;">
              <p style="margin:0 0 4px 0; color:#1a1a1a; font-size:12px; font-weight:600;">GL BAJAJ INSTITUTE OF TECHNOLOGY AND MANAGEMENT</p>
              <p style="margin:0 0 8px 0; color:#9aa0a6; font-size:12px;">Report It</p>
              <p style="margin:0; color:#b0b5bb; font-size:11px;">This is an automated message. Please do not reply to this email.</p>
            </td>
          </tr>
 
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`, otp)
}

const (
	smtpHost    = "smtp.gmail.com"
	smtpPort    = "587"
	senderEmail = "vedantvisoliya@gmail.com"
)

func SendOTPEmail(toEmail string, otp string, appPassword string) error {
	htmlBody := buildOTPEmailHTML(otp) // reuse the same HTML builder from before

	subject := fmt.Sprintf("Subject: %s is your Report It verification code\r\n", otp)
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	fromHeader := fmt.Sprintf("From: Report It <%s>\r\n", senderEmail)
	toHeader := fmt.Sprintf("To: %s\r\n", toEmail)

	message := []byte(fromHeader + toHeader + subject + mime + "\r\n" + htmlBody)

	auth := smtp.PlainAuth("", senderEmail, appPassword, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// using resend otp service
// func SendOTPEmail(email string, otp string, resendAPIKey string) (*resend.SendEmailResponse, error) {
// 	client := resend.NewClient(resendAPIKey)

// 	htmlBody := buildOTPEmailHTML(otp)

// 	params := &resend.SendEmailRequest{
// 		To:      []string{email},
// 		From:    "Report It <vedantvisoliya@gmail.com>",
// 		Subject: fmt.Sprintf("%s is your Report It verification code", otp),
// 		Text:    fmt.Sprintf("Your Report It verification code is %s", otp),
// 		Html:    htmlBody,
// 	}

// 	sent, err := client.Emails.Send(params)
// 	if err != nil {
// 		return &resend.SendEmailResponse{}, err
// 	}

// 	return sent, nil
// }

package email

import "fmt"

// GetVerificationEmailContent returns the HTML content for the verification email.
func GetVerificationEmailContent(name, verificationLink string) string {
	return fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Verify your email address</h2>
			<p>Hi %s,</p>
			<p>Thanks for signing up! Please verify your email address by clicking the link below:</p>
			<p>
				<a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
			</p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 24 hours.</p>
		</div>
	`, name, verificationLink, verificationLink)
}

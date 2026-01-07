package email

import "fmt"

// GetPasswordResetEmailContent returns the HTML content for the password reset email.
func GetPasswordResetEmailContent(name, resetLink string) string {
	return fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Reset your password</h2>
			<p>Hi %s,</p>
			<p>We received a request to reset your password. If you didn't make this request, you can safely ignore this email.</p>
			<p>To reset your password, click the link below:</p>
			<p>
				<a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
			</p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 1 hour.</p>
		</div>
	`, name, resetLink, resetLink)
}

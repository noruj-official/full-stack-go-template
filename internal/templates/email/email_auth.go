package email

import "fmt"

// GetEmailAuthEmailContent returns the HTML content for the email authentication email.
func GetEmailAuthEmailContent(appName, link string) string {
	return fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Sign in to %s</h2>
			<p>Hello,</p>
			<p>Click the button below to sign in to your account. This link will expire in 15 minutes.</p>
			<p>
				<a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Sign in</a>
			</p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>If you didn't request this email, you can safely ignore it.</p>
		</div>
	`, appName, link, link)
}

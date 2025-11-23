package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Service defines the email service interface
type Service interface {
	SendWelcomeEmail(to, firstName, dealershipName string) error
	SendTrialEndingEmail(to, firstName, dealershipName string, daysLeft int) error
	SendPaymentSuccessEmail(to, firstName, amount float64, planName string) error
	SendPasswordResetEmail(to, firstName, resetLink string) error
	SendInviteEmail(to, firstName, inviterName, dealershipName, inviteLink string) error
}

// BrevoService implements email service using Brevo API
type BrevoService struct {
	apiKey      string
	senderEmail string
	senderName  string
	client      *http.Client
}

// NewBrevoService creates a new Brevo email service
func NewBrevoService() *BrevoService {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := getEnv("BREVO_SENDER_EMAIL", "noreply@carflow.com")
	senderName := getEnv("BREVO_SENDER_NAME", "CarFlow")

	return &BrevoService{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		senderName:  senderName,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BrevoEmailRequest represents the Brevo API request structure
type BrevoEmailRequest struct {
	Sender  Sender       `json:"sender"`
	To      []Recipient  `json:"to"`
	Subject string       `json:"subject"`
	HtmlContent string   `json:"htmlContent"`
}

type Sender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Recipient struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// sendEmail sends an email via Brevo API
func (s *BrevoService) sendEmail(to, toName, subject, htmlContent string) error {
	if s.apiKey == "" {
		log.Printf("⚠️  Brevo API key not configured, email would be sent to: %s", to)
		log.Printf("   Subject: %s", subject)
		return nil // Don't fail if API key is not set (development mode)
	}

	reqBody := BrevoEmailRequest{
		Sender: Sender{
			Name:  s.senderName,
			Email: s.senderEmail,
		},
		To: []Recipient{
			{
				Name:  toName,
				Email: to,
			},
		},
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo API error (status %d): %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Email sent to %s: %s", to, subject)
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (s *BrevoService) SendWelcomeEmail(to, firstName, dealershipName string) error {
	subject := fmt.Sprintf("Welcome to CarFlow, %s! 🚗", firstName)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚗 Welcome to CarFlow!</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            <p>Welcome to CarFlow! We're excited to have <strong>%s</strong> on board.</p>

            <p>Your account has been created successfully and you're ready to start managing your vehicle inventory in the cloud.</p>

            <h3>🎉 What's Included:</h3>
            <ul>
                <li>✅ 14-day free trial (no credit card required)</li>
                <li>✅ Complete vehicle inventory management</li>
                <li>✅ Advanced filtering and search</li>
                <li>✅ RESTful API access</li>
                <li>✅ Team collaboration features</li>
            </ul>

            <h3>🚀 Get Started:</h3>
            <p>Log in to your dashboard and add your first vehicle. It only takes a minute!</p>

            <a href="http://localhost:3000/login" class="button">Go to Dashboard →</a>

            <p>If you have any questions, just reply to this email. We're here to help!</p>

            <p>Best regards,<br>
            The CarFlow Team</p>
        </div>
        <div class="footer">
            <p>CarFlow - Fleet Management SaaS</p>
            <p>© 2025 CarFlow. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, firstName, dealershipName)

	return s.sendEmail(to, firstName, subject, html)
}

// SendTrialEndingEmail sends a notification when trial is ending
func (s *BrevoService) SendTrialEndingEmail(to, firstName, dealershipName string, daysLeft int) error {
	subject := fmt.Sprintf("Your CarFlow trial ends in %d days", daysLeft)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #F59E0B; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .pricing { background: #f9fafb; padding: 15px; border-radius: 6px; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ Trial Ending Soon</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            <p>Your CarFlow trial for <strong>%s</strong> ends in <strong>%d days</strong>.</p>

            <p>We hope you've enjoyed using CarFlow to manage your vehicle inventory! To continue using all features, choose a plan that works for you.</p>

            <div class="pricing">
                <h3>💼 Choose Your Plan:</h3>
                <ul>
                    <li><strong>Starter:</strong> $29/mo - Up to 100 vehicles</li>
                    <li><strong>Professional:</strong> $99/mo - Up to 1,000 vehicles</li>
                    <li><strong>Enterprise:</strong> $299/mo - Unlimited vehicles</li>
                </ul>
            </div>

            <a href="http://localhost:3000/billing" class="button">Choose Your Plan →</a>

            <p><em>No action needed if you want to cancel - your trial will simply expire.</em></p>

            <p>Questions? Reply to this email and we'll help!</p>

            <p>Best regards,<br>
            The CarFlow Team</p>
        </div>
        <div class="footer">
            <p>CarFlow - Fleet Management SaaS</p>
        </div>
    </div>
</body>
</html>
`, firstName, dealershipName, daysLeft)

	return s.sendEmail(to, firstName, subject, html)
}

// SendPaymentSuccessEmail sends a confirmation after successful payment
func (s *BrevoService) SendPaymentSuccessEmail(to, firstName string, amount float64, planName string) error {
	subject := "Payment received - Thank you! 💳"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .invoice { background: #f9fafb; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Payment Successful!</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            <p>Thank you for your payment! Your subscription is now active.</p>

            <div class="invoice">
                <h3>Payment Details:</h3>
                <p><strong>Plan:</strong> %s</p>
                <p><strong>Amount:</strong> $%.2f</p>
                <p><strong>Date:</strong> %s</p>
            </div>

            <p>You can view your invoice and manage your subscription in your dashboard.</p>

            <p>Thank you for choosing CarFlow!</p>

            <p>Best regards,<br>
            The CarFlow Team</p>
        </div>
        <div class="footer">
            <p>CarFlow - Fleet Management SaaS</p>
        </div>
    </div>
</body>
</html>
`, firstName, planName, amount, time.Now().Format("January 2, 2006"))

	return s.sendEmail(to, firstName, subject, html)
}

// SendPasswordResetEmail sends a password reset link
func (s *BrevoService) SendPasswordResetEmail(to, firstName, resetLink string) error {
	subject := "Reset your CarFlow password"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .warning { background: #FEF3C7; padding: 15px; border-left: 4px solid #F59E0B; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Password Reset</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            <p>We received a request to reset your CarFlow password.</p>

            <p>Click the button below to reset your password:</p>

            <a href="%s" class="button">Reset Password →</a>

            <p>Or copy and paste this link into your browser:<br>
            <code>%s</code></p>

            <div class="warning">
                <strong>⚠️ Security Notice:</strong><br>
                This link will expire in 1 hour. If you didn't request this reset, please ignore this email.
            </div>

            <p>Best regards,<br>
            The CarFlow Team</p>
        </div>
        <div class="footer">
            <p>CarFlow - Fleet Management SaaS</p>
        </div>
    </div>
</body>
</html>
`, firstName, resetLink, resetLink)

	return s.sendEmail(to, firstName, subject, html)
}

// SendInviteEmail sends a team member invitation
func (s *BrevoService) SendInviteEmail(to, firstName, inviterName, dealershipName, inviteLink string) error {
	subject := fmt.Sprintf("%s invited you to join %s on CarFlow", inviterName, dealershipName)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 Team Invitation</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            <p><strong>%s</strong> has invited you to join <strong>%s</strong> on CarFlow!</p>

            <p>CarFlow is a modern fleet management platform that helps dealerships manage their vehicle inventory in the cloud.</p>

            <p>Accept the invitation to get started:</p>

            <a href="%s" class="button">Accept Invitation →</a>

            <p>Looking forward to having you on the team!</p>

            <p>Best regards,<br>
            The CarFlow Team</p>
        </div>
        <div class="footer">
            <p>CarFlow - Fleet Management SaaS</p>
        </div>
    </div>
</body>
</html>
`, firstName, inviterName, dealershipName, inviteLink)

	return s.sendEmail(to, firstName, subject, html)
}

// MockEmailService implements email service for testing (logs only)
type MockEmailService struct{}

// NewMockEmailService creates a mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendWelcomeEmail(to, firstName, dealershipName string) error {
	log.Printf("📧 [MOCK] Welcome email to %s (%s) for %s", to, firstName, dealershipName)
	return nil
}

func (m *MockEmailService) SendTrialEndingEmail(to, firstName, dealershipName string, daysLeft int) error {
	log.Printf("📧 [MOCK] Trial ending email to %s (%s) - %d days left", to, firstName, daysLeft)
	return nil
}

func (m *MockEmailService) SendPaymentSuccessEmail(to, firstName string, amount float64, planName string) error {
	log.Printf("📧 [MOCK] Payment success email to %s (%s) - $%.2f for %s", to, firstName, amount, planName)
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(to, firstName, resetLink string) error {
	log.Printf("📧 [MOCK] Password reset email to %s (%s) - Link: %s", to, firstName, resetLink)
	return nil
}

func (m *MockEmailService) SendInviteEmail(to, firstName, inviterName, dealershipName, inviteLink string) error {
	log.Printf("📧 [MOCK] Invite email to %s (%s) from %s for %s", to, firstName, inviterName, dealershipName)
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"time"
)

// NotificationService handles email and alert delivery
type NotificationService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	fromAddress  string
	enabled      bool
}

// NotificationPolicy defines when to trigger notifications
type NotificationPolicy struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Repository      string        `json:"repository"`
	TriggerOn       []string      `json:"triggerOn"` // ["critical", "high", "ai_generated"]
	Recipients      []string      `json:"recipients"`
	Enabled         bool          `json:"enabled"`
	CreatedAt       time.Time     `json:"createdAt"`
	LastNotified    *time.Time    `json:"lastNotified"`
	NotificationGap time.Duration `json:"notificationGap"` // Cooldown period
}

// NotificationEvent represents an event that may trigger notifications
type NotificationEvent struct {
	Type       string // "analysis_complete", "critical_risk", "policy_violation"
	Repository string
	PRNumber   int
	Analysis   *AnalyzeResponse
	Severity   string // "critical", "high", "medium", "low"
	Message    string
}

// NewNotificationService initializes the email service
func NewNotificationService(smtpHost, smtpPort, smtpUser, smtpPassword, fromAddress string) *NotificationService {
	enabled := smtpHost != "" && smtpUser != ""
	return &NotificationService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		enabled:      enabled,
	}
}

// ProcessEvent checks if notification should be sent and sends it
func (ns *NotificationService) ProcessEvent(ctx context.Context, event *NotificationEvent, policy *NotificationPolicy) error {
	if !ns.enabled || !policy.Enabled {
		return nil
	}

	// Check if event matches policy triggers
	if !ns.shouldNotify(event, policy) {
		slog.Debug("notification suppressed by policy", "repo", policy.Repository, "trigger_on", policy.TriggerOn)
		return nil
	}

	// Check cooldown period
	if policy.LastNotified != nil {
		timeSinceLastNotification := time.Now().Sub(*policy.LastNotified)
		if timeSinceLastNotification < policy.NotificationGap {
			slog.Info("notification suppressed by cooldown", "repo", policy.Repository, "next_allowed", policy.LastNotified.Add(policy.NotificationGap))
			return nil
		}
	}

	// Build email
	subject, body := ns.buildEmailContent(event, policy)

	// Send to all recipients
	for _, recipient := range policy.Recipients {
		if err := ns.sendEmail(recipient, subject, body); err != nil {
			slog.Error("failed to send notification email", "recipient", recipient, "error", err)
			continue
		}
		slog.Info("notification sent", "recipient", recipient, "repo", policy.Repository)
	}

	return nil
}

// shouldNotify checks if event type matches policy triggers
func (ns *NotificationService) shouldNotify(event *NotificationEvent, policy *NotificationPolicy) bool {
	for _, trigger := range policy.TriggerOn {
		switch trigger {
		case "critical":
			if event.Analysis != nil && event.Analysis.Critical > 0 {
				return true
			}
		case "high":
			if event.Analysis != nil && event.Analysis.High > 0 {
				return true
			}
		case "ai_generated":
			if event.Analysis != nil && event.Analysis.Recommendation == "BLOCK" {
				return true
			}
		case "policy_violation":
			if event.Type == "policy_violation" {
				return true
			}
		}
	}
	return false
}

// buildEmailContent generates email subject and body
func (ns *NotificationService) buildEmailContent(event *NotificationEvent, policy *NotificationPolicy) (string, string) {
	subject := fmt.Sprintf("[Tribunal Alert] %s - PR #%d", event.Repository, event.PRNumber)

	if event.Analysis != nil && event.Analysis.Critical > 0 {
		subject = fmt.Sprintf("[CRITICAL] %s", subject)
	} else if event.Analysis != nil && event.Analysis.High > 0 {
		subject = fmt.Sprintf("[HIGH] %s", subject)
	}

	html := ns.buildHTMLEmail(event, policy)
	return subject, html
}

// buildHTMLEmail generates HTML email body
func (ns *NotificationService) buildHTMLEmail(event *NotificationEvent, policy *NotificationPolicy) string {
	analysis := event.Analysis
	if analysis == nil {
		return fmt.Sprintf("<p>%s</p>", event.Message)
	}

	riskSummary := fmt.Sprintf(`
    <table style="width: 100%%; border-collapse: collapse;">
      <tr style="background: #f3f4f6;">
        <td style="padding: 10px; border: 1px solid #e5e7eb; color: #dc2626; font-weight: bold;">
          Critical: %d
        </td>
        <td style="padding: 10px; border: 1px solid #e5e7eb; color: #ea580c; font-weight: bold;">
          High: %d
        </td>
        <td style="padding: 10px; border: 1px solid #e5e7eb; color: #eab308; font-weight: bold;">
          Medium: %d
        </td>
        <td style="padding: 10px; border: 1px solid #e5e7eb; color: #16a34a; font-weight: bold;">
          Low: %d
        </td>
      </tr>
    </table>
  `, analysis.Critical, analysis.High, analysis.Medium, analysis.Low)

	recommendation := ""
	if analysis.Recommendation == "BLOCK" {
		recommendation = `<p style="color: #dc2626; font-weight: bold;">⚠️ RECOMMENDATION: BLOCK</p>`
	} else if analysis.Recommendation == "REVIEW_REQUIRED" {
		recommendation = `<p style="color: #ea580c; font-weight: bold;">⚠️ RECOMMENDATION: REVIEW REQUIRED</p>`
	}

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.6; color: #1f2937; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1f2937; color: white; padding: 20px; border-radius: 8px; }
        .content { margin: 20px 0; }
        .footer { color: #9ca3af; font-size: 12px; margin-top: 20px; border-top: 1px solid #e5e7eb; padding-top: 20px; }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <h2>🛡️ Tribunal Security Alert</h2>
          <p>Policy: %s</p>
        </div>

        <div class="content">
          <h3>Pull Request Analysis</h3>
          <p><strong>Repository:</strong> %s</p>
          <p><strong>PR Number:</strong> #%d</p>
          <p><strong>Files Analyzed:</strong> %d</p>
          <p><strong>AI Generated:</strong> %d</p>

          <h3>Risk Summary</h3>
          %s

          %s

          <h3>Summary</h3>
          <p>Analysis recommendation: <strong>%s</strong></p>

          <h3>Action Required</h3>
          <p>Please review this PR and take appropriate action based on the risk level and recommendation.</p>
        </div>

        <div class="footer">
          <p>This is an automated notification from Tribunal Security Audit System.</p>
          <p>Generated: %s</p>
        </div>
      </div>
    </body>
    </html>
  `,
		policy.Name,
		event.Repository,
		event.PRNumber,
		analysis.TotalFiles,
		analysis.AIGenerated,
		riskSummary,
		recommendation,
		analysis.Recommendation,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return html
}

// sendEmail sends an email via SMTP
func (ns *NotificationService) sendEmail(to, subject, body string) error {
	if !ns.enabled {
		slog.Warn("email service not configured, skipping send")
		return nil
	}

	auth := smtp.PlainAuth("", ns.smtpUser, ns.smtpPassword, ns.smtpHost)
	addr := fmt.Sprintf("%s:%s", ns.smtpHost, ns.smtpPort)

	// Email headers
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		ns.fromAddress,
		to,
		subject,
	)

	fullMessage := headers + body

	err := smtp.SendMail(addr, auth, ns.fromAddress, []string{to}, []byte(fullMessage))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendBulkNotifications sends notifications to multiple policies
func (ns *NotificationService) SendBulkNotifications(ctx context.Context, event *NotificationEvent, policies []*NotificationPolicy) error {
	for _, policy := range policies {
		// Filter policies by repository
		if policy.Repository != "" && policy.Repository != event.Repository {
			continue
		}

		if err := ns.ProcessEvent(ctx, event, policy); err != nil {
			slog.Error("failed to process notification", "policy", policy.ID, "error", err)
		}
	}

	return nil
}

// CreateDefaultPolicies returns sensible default notification policies
func CreateDefaultPolicies(repo string) []*NotificationPolicy {
	return []*NotificationPolicy{
		{
			ID:              "policy_critical_alerts",
			Name:            "Critical Risk Alerts",
			Repository:      repo,
			TriggerOn:       []string{"critical", "policy_violation"},
			Recipients:      []string{}, // Must be configured by user
			Enabled:         true,
			CreatedAt:       time.Now(),
			NotificationGap: 30 * time.Minute,
		},
		{
			ID:              "policy_ai_generated",
			Name:            "AI-Generated Code Detection",
			Repository:      repo,
			TriggerOn:       []string{"ai_generated"},
			Recipients:      []string{},
			Enabled:         true,
			CreatedAt:       time.Now(),
			NotificationGap: 1 * time.Hour,
		},
		{
			ID:              "policy_high_risk",
			Name:            "High Risk Findings",
			Repository:      repo,
			TriggerOn:       []string{"high"},
			Recipients:      []string{},
			Enabled:         false, // Disabled by default to reduce noise
			CreatedAt:       time.Now(),
			NotificationGap: 2 * time.Hour,
		},
	}
}

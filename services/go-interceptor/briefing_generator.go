package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// EnhancedBriefingGenerator creates detailed context-aware briefings.
type EnhancedBriefingGenerator struct{}

// NewEnhancedBriefingGenerator creates a new briefing generator.
func NewEnhancedBriefingGenerator() *EnhancedBriefingGenerator {
	return &EnhancedBriefingGenerator{}
}

// GenerateEnhancedBriefing creates a context-aware briefing with incident correlation.
func (g *EnhancedBriefingGenerator) GenerateEnhancedBriefing(resp *AnalyzeResponse, repoCtx *RepositoryContext) string {
	var buf bytes.Buffer

	// Title
	buf.WriteString(fmt.Sprintf("# TRIBUNAL Context Briefing: PR #%d\n\n", resp.PRNumber))
	buf.WriteString(fmt.Sprintf("**Repository**: %s\n\n", resp.Repository))

	// ============================================================================
	// SECTION 1: Repository Context (if available)
	// ============================================================================
	if repoCtx != nil && repoCtx.ContextRelevanceScore > 0 {
		buf.WriteString("## 📋 Repository Context\n\n")

		// Service dependencies
		if len(repoCtx.ServiceDependencies) > 0 {
			buf.WriteString("**Critical Dependencies**:\n")
			for _, dep := range repoCtx.ServiceDependencies {
				icon := "🔗"
				if dep.Critical {
					icon = "🔴"
				}
				buf.WriteString(fmt.Sprintf("- %s %s (%s)\n", icon, dep.Dependency, dep.Type))
			}
			buf.WriteString("\n")
		}

		// Recent incidents
		if len(repoCtx.Incidents) > 0 {
			buf.WriteString("**Historical Incidents** (⚠️ Recent issues in this codebase):\n")
			for i, incident := range repoCtx.Incidents {
				if i >= 3 {
					break
				}
				severity := "🔴"
				if incident.Severity == "HIGH" {
					severity = "🟠"
				} else if incident.Severity == "MEDIUM" {
					severity = "🟡"
				}
				buf.WriteString(fmt.Sprintf("- %s **%s** (%s): %s\n", severity, incident.Title, incident.DateOccurred, incident.RootCause))
			}
			buf.WriteString("\n")
		}

		// Deployment stability
		if len(repoCtx.DeploymentHistory) > 0 {
			stableDeployments := 0
			problematicDeployments := 0

			for _, event := range repoCtx.DeploymentHistory {
				if event.Incidents == 0 {
					stableDeployments++
				} else {
					problematicDeployments++
				}
			}

			buf.WriteString("**Deployment Stability**:\n")
			buf.WriteString(fmt.Sprintf("- Stable releases: %d / %d\n", stableDeployments, len(repoCtx.DeploymentHistory)))
			if problematicDeployments > 0 {
				buf.WriteString(fmt.Sprintf("- ⚠️ Releases with incidents: %d\n", problematicDeployments))
			}
			buf.WriteString("\n")
		}
	}

	// ============================================================================
	// SECTION 2: Executive Summary with Risk Assessment
	// ============================================================================
	buf.WriteString("## 🎯 Executive Summary\n\n")

	riskEmoji := "✅"
	if resp.Recommendation == "BLOCK" {
		riskEmoji = "🚨"
	} else if resp.Recommendation == "REVIEW_REQUIRED" {
		riskEmoji = "⚠️"
	}

	buf.WriteString(fmt.Sprintf("%s **RECOMMENDATION: %s**\n\n", riskEmoji, resp.Recommendation))

	// Risk distribution
	buf.WriteString("**Risk Distribution**:\n")
	buf.WriteString(fmt.Sprintf("- 🔴 Critical: %d / %d files\n", resp.Critical, resp.TotalFiles))
	buf.WriteString(fmt.Sprintf("- 🟠 High: %d / %d files\n", resp.High, resp.TotalFiles))
	buf.WriteString(fmt.Sprintf("- 🟡 Medium: %d / %d files\n", resp.Medium, resp.TotalFiles))
	buf.WriteString(fmt.Sprintf("- 🟢 Low: %d / %d files\n\n", resp.Low, resp.TotalFiles))

	// AI detection
	if resp.AIGenerated > 0 {
		aiPercent := float64(resp.AIGenerated) / float64(resp.TotalFiles) * 100
		buf.WriteString(fmt.Sprintf("⚠️ **AI-Generation Alert**: %d files (%.1f%%) show AI-generation markers\n\n", resp.AIGenerated, aiPercent))
	}

	// ============================================================================
	// SECTION 3: High-Risk Files with Context
	// ============================================================================
	if resp.Critical > 0 || resp.High > 0 {
		buf.WriteString("## 🔴 Critical Findings\n\n")

		for _, f := range resp.Files {
			if f.RiskLevel == "CRITICAL" || f.RiskLevel == "HIGH" {
				buf.WriteString(fmt.Sprintf("### %s\n\n", f.Path))

				// Risk indicators
				riskIcon := "🟠"
				if f.RiskLevel == "CRITICAL" {
					riskIcon = "🔴"
				}

				buf.WriteString(fmt.Sprintf("**%s Risk Level**: %s\n", riskIcon, f.RiskLevel))
				buf.WriteString(fmt.Sprintf("**AI Likelihood**: %.0f%% (Confidence: %.0f%%)\n\n", f.AIScore*100, f.Confidence*100))

				// Signal breakdown
				buf.WriteString("**Detection Signals**:\n")
				buf.WriteString(fmt.Sprintf("- 🎨 Style anomaly: %.2f (verbose identifiers, unusual patterns)\n", f.Signals.Style))
				buf.WriteString(fmt.Sprintf("- 🔍 Pattern markers: %.2f (AI generation keywords, explicit markers)\n", f.Signals.Pattern))
				buf.WriteString(fmt.Sprintf("- ⚠️ Risk keywords: %.2f (critical operations, dangerous patterns)\n\n", f.Signals.Risk))

				// Summary
				buf.WriteString(fmt.Sprintf("**Analysis Summary**:\n%s\n\n", f.Summary))

				// Contextual warnings based on repository context
				if repoCtx != nil && len(repoCtx.Incidents) > 0 {
					// Check if this file might affect critical dependencies
					relatedIncidents := findRelatedIncidents(f.Path, repoCtx.Incidents)
					if len(relatedIncidents) > 0 {
						buf.WriteString("**⚠️ Incident History Alert**:\n")
						for _, inc := range relatedIncidents {
							buf.WriteString(fmt.Sprintf("- Previous issue: %s (%s)\n", inc.Title, inc.Severity))
						}
						buf.WriteString("\n")
					}
				}

				// Suggested fix
				if f.SuggestedFix != "" {
					buf.WriteString("**🛠️ Suggested Remediation**:\n")
					buf.WriteString("```go\n")
					buf.WriteString(f.SuggestedFix)
					buf.WriteString("\n```\n\n")
				}

				buf.WriteString("---\n\n")
			}
		}
	}

	// ============================================================================
	// SECTION 4: Medium-Risk Files
	// ============================================================================
	if resp.Medium > 0 {
		buf.WriteString("## 🟡 Medium-Risk Files (Review Required)\n\n")

		for _, f := range resp.Files {
			if f.RiskLevel == "MEDIUM" {
				buf.WriteString(fmt.Sprintf("- **%s**: AI Score %.0f%%, Confidence %.0f%%\n", f.Path, f.AIScore*100, f.Confidence*100))
			}
		}
		buf.WriteString("\n")
	}

	// ============================================================================
	// SECTION 5: Recommendation & Next Steps
	// ============================================================================
	buf.WriteString("## 📌 Recommendation & Next Steps\n\n")

	switch resp.Recommendation {
	case "APPROVE":
		buf.WriteString("✅ **GREEN LIGHT**: Changes are low-risk and align with established patterns.\n\n")
		buf.WriteString("**Action**: Safe to merge. No contextual blindspots detected.\n\n")

	case "REVIEW_REQUIRED":
		buf.WriteString("⚠️ **REQUEST CHANGES**: This PR requires additional human review.\n\n")
		buf.WriteString("**Reasons**:\n")
		if resp.AIGenerated > 0 {
			buf.WriteString(fmt.Sprintf("- %d file(s) show AI-generation markers\n", resp.AIGenerated))
		}
		if resp.High > 0 {
			buf.WriteString(fmt.Sprintf("- %d high-risk file(s) detected\n", resp.High))
		}
		buf.WriteString("\n")
		buf.WriteString("**Action**: Review flagged sections for:\n")
		buf.WriteString("- Missing error handling\n")
		buf.WriteString("- Database consistency issues\n")
		buf.WriteString("- Idempotency violations\n")
		buf.WriteString("- Scale/performance impacts\n")
		buf.WriteString("- Security boundary violations\n\n")

	case "BLOCK":
		buf.WriteString("🛑 **BLOCK MERGE**: Critical issues detected.\n\n")
		buf.WriteString(fmt.Sprintf("**Blocking Issues**: %d critical-risk files\n\n", resp.Critical))
		buf.WriteString("**Required Actions**:\n")
		buf.WriteString("1. Review suggested fixes above\n")
		buf.WriteString("2. Consult with team about:\n")
		buf.WriteString("   - Service topology impact\n")
		buf.WriteString("   - Incident correlation warnings\n")
		buf.WriteString("   - Database/cache consistency\n")
		buf.WriteString("3. Resubmit with revisions\n\n")
	}

	// ============================================================================
	// SECTION 6: Debug Info
	// ============================================================================
	buf.WriteString("---\n\n")
	buf.WriteString("**TRIBUNAL Analysis Metadata**\n\n")
	buf.WriteString(fmt.Sprintf("- **Files Analyzed**: %d\n", resp.TotalFiles))
	buf.WriteString(fmt.Sprintf("- **Analysis Time**: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))
	if repoCtx != nil {
		buf.WriteString(fmt.Sprintf("- **Context Relevance**: %.1f%%\n", repoCtx.ContextRelevanceScore*100))
		buf.WriteString(fmt.Sprintf("- **Dependencies Tracked**: %d\n", len(repoCtx.ServiceDependencies)))
		buf.WriteString(fmt.Sprintf("- **Incident History**: %d known issues\n", len(repoCtx.Incidents)))
	}
	buf.WriteString("\n_Generated by TRIBUNAL AI Context Analyzer_\n")

	return buf.String()
}

// findRelatedIncidents looks for incidents that might be related to a file change.
func findRelatedIncidents(filePath string, incidents []IncidentPattern) []IncidentPattern {
	var related []IncidentPattern

	filePathLower := strings.ToLower(filePath)

	for _, incident := range incidents {
		// Check if incident mentions this file or related code patterns
		for _, code := range incident.AffectedCode {
			if strings.Contains(strings.ToLower(code), strings.ToLower(filePath)) ||
				strings.Contains(filePathLower, strings.ToLower(code)) {
				related = append(related, incident)
				break
			}
		}

		// Also check if the file path contains keywords from incident description
		if strings.Contains(filePathLower, "db") && strings.Contains(strings.ToLower(incident.Description), "database") {
			related = append(related, incident)
			break
		}
		if strings.Contains(filePathLower, "migration") && strings.Contains(strings.ToLower(incident.Description), "migration") {
			related = append(related, incident)
			break
		}
	}

	return related
}

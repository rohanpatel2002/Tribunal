package main

import "time"

type ChangedFile struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Patch  string `json:"patch"`
}

type AnalyzeRequest struct {
	Repository string        `json:"repository"`
	PRNumber   int           `json:"prNumber"`
	Files      []ChangedFile `json:"files"`
}

type SignalBreakdown struct {
	Style   float64 `json:"style"`
	Pattern float64 `json:"pattern"`
	Risk    float64 `json:"risk"`
}

type FileAnalysis struct {
	Path          string          `json:"path"`
	AIScore       float64         `json:"aiScore"`
	IsAIGenerated bool            `json:"isAIGenerated"`
	Confidence    float64         `json:"confidence"`
	Signals       SignalBreakdown `json:"signals"`
	RiskLevel     string          `json:"riskLevel"`
	Summary       string          `json:"summary"`
	SuggestedFix  string          `json:"suggestedFix,omitempty"`
}

type AnalysisSummary struct {
	TotalFiles     int    `json:"totalFiles"`
	AIGenerated    int    `json:"aiGenerated"`
	Critical       int    `json:"critical"`
	High           int    `json:"high"`
	Medium         int    `json:"medium"`
	Low            int    `json:"low"`
	Recommendation string `json:"recommendation"`
}

type AnalyzeResponse struct {
	ID             string          `json:"id,omitempty"`
	Repository     string          `json:"repository"`
	PRNumber       int             `json:"prNumber"`
	Recommendation string          `json:"recommendation"`
	TotalFiles     int             `json:"totalFiles"`
	AIGenerated    int             `json:"aiGenerated"`
	Critical       int             `json:"critical"`
	High           int             `json:"high"`
	Medium         int             `json:"medium"`
	Low            int             `json:"low"`
	Files          []FileAnalysis  `json:"files,omitempty"`
	CreatedAt      time.Time       `json:"createdAt,omitempty"`
}

type AuditSummary struct {
	Repository     string  `json:"repository"`
	TotalPRs       int     `json:"totalPRs"`
	TotalFiles     int     `json:"totalFiles"`
	AIGeneratedPRs int     `json:"aiGeneratedPRs"`
	CriticalRisks  int     `json:"criticalRisks"`
	HighRisks      int     `json:"highRisks"`
	AverageAIScore float64 `json:"averageAIScore"`
}

type PRAnalysisRecord struct {
	ID             string `json:"id"`
	Repository     string `json:"repository"`
	PRNumber       int    `json:"prNumber"`
	Recommendation string `json:"recommendation"`
	TotalFiles     int    `json:"totalFiles"`
	AIGenerated    int    `json:"aiGenerated"`
	Critical       int    `json:"critical"`
	High           int    `json:"high"`
	Medium         int    `json:"medium"`
	Low            int    `json:"low"`
}

// SecurityPolicy defines a security enforcement rule
type SecurityPolicy struct {
	ID                  string                 `json:"id,omitempty"`
	Repository          string                 `json:"repository"`
	PolicyName          string                 `json:"policyName"`
	PolicyType          string                 `json:"policyType"` // AI_DETECTION, VULNERABILITY_SCAN, CODE_STYLE, COMPLIANCE
	Description         string                 `json:"description"`
	Rules               map[string]interface{} `json:"rules"`
	IsActive            bool                   `json:"isActive"`
	SeverityThreshold   string                 `json:"severityThreshold"` // LOW, MEDIUM, HIGH, CRITICAL
	CreatedBy           string                 `json:"createdBy,omitempty"`
	CreatedAt           time.Time              `json:"createdAt,omitempty"`
	UpdatedAt           time.Time              `json:"updatedAt,omitempty"`
}

// SecurityEvent represents an audit trail event
type SecurityEvent struct {
	ID             string                 `json:"id"`
	EventType      string                 `json:"eventType"` // POLICY_CREATED, POLICY_UPDATED, ANALYSIS_COMPLETED, etc.
	Repository     string                 `json:"repository,omitempty"`
	Actor          string                 `json:"actor"`
	ActionDetails  map[string]interface{} `json:"actionDetails,omitempty"`
	Severity       string                 `json:"severity"` // INFO, WARNING, ERROR, CRITICAL
	CreatedAt      time.Time              `json:"createdAt"`
}

// APIKey represents a security credential
type APIKey struct {
	ID         string    `json:"id"`
	KeyName    string    `json:"keyName"`
	Repository string    `json:"repository,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	IsActive   bool      `json:"isActive"`
}

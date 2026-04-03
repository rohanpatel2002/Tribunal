package main

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
	Repository string          `json:"repository"`
	PRNumber   int             `json:"prNumber"`
	Summary    AnalysisSummary `json:"summary"`
	Results    []FileAnalysis  `json:"results"`
}

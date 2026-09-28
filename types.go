package altirs

// Severity level for a detected violation.
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// CheckName identifies which guardrail check to run.
type CheckName string

const (
	CheckContentSafety   CheckName = "content_safety"
	CheckPII             CheckName = "pii"
	CheckPromptInjection CheckName = "prompt_injection"
	CheckHallucination   CheckName = "hallucination"
)

// Violation represents a single violation detected by a guardrail check.
type Violation struct {
	Type        string   `json:"type"`
	Category    string   `json:"category"`
	Severity    Severity `json:"severity"`
	Description string   `json:"description"`
}

// PiiOptions configures PII detection behavior.
type PiiOptions struct {
	Mask *bool `json:"mask,omitempty"`
}

// HallucinationOptions configures hallucination detection behavior.
type HallucinationOptions struct {
	Context *string `json:"context,omitempty"`
}

// CheckOptions holds per-check configuration.
type CheckOptions struct {
	PII           *PiiOptions           `json:"pii,omitempty"`
	Hallucination *HallucinationOptions `json:"hallucination,omitempty"`
}

// CheckPayload is the payload for a single guardrail check.
type CheckPayload struct {
	Input    *string      `json:"input,omitempty"`
	Response *string      `json:"response,omitempty"`
	Checks   []CheckName  `json:"checks,omitempty"`
	Options  *CheckOptions `json:"options,omitempty"`
}

// CheckResponse is the response from a single guardrail check.
type CheckResponse struct {
	Safe           bool        `json:"safe"`
	Blocked        bool        `json:"blocked"`
	Score          float64     `json:"score"`
	Violations     []Violation `json:"violations"`
	MaskedInput    *string     `json:"maskedInput,omitempty"`
	MaskedResponse *string     `json:"maskedResponse,omitempty"`
	Recommendation *string     `json:"recommendation,omitempty"`
	RequestID      string      `json:"requestId"`
	LatencyMs      int         `json:"latencyMs"`
}

// BatchPayload is the payload for a batch guardrail check.
type BatchPayload struct {
	Items []CheckPayload `json:"items"`
}

// BatchResultItem is an individual result within a batch response.
type BatchResultItem struct {
	Index          int         `json:"index"`
	Safe           bool        `json:"safe"`
	Blocked        bool        `json:"blocked"`
	Score          float64     `json:"score"`
	Violations     []Violation `json:"violations"`
	MaskedInput    *string     `json:"maskedInput,omitempty"`
	MaskedResponse *string     `json:"maskedResponse,omitempty"`
	Recommendation *string     `json:"recommendation,omitempty"`
	LatencyMs      int         `json:"latencyMs"`
}

// BatchResponse is the response from a batch guardrail check.
type BatchResponse struct {
	Results   []BatchResultItem `json:"results"`
	Total     int               `json:"total"`
	RequestID string            `json:"requestId"`
	LatencyMs int               `json:"latencyMs"`
}

// ErrorResponse is the error payload returned by the API.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

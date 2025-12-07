package contract

import (
	"context"
)

// ConfluenceClient defines Confluence API operations
type ConfluenceClient interface {
	// Page operations
	CreatePage(ctx context.Context, page *ConfluencePage) (*ConfluencePage, error)
	UpdatePage(ctx context.Context, pageID string, page *ConfluencePageUpdate) (*ConfluencePage, error)
	GetPage(ctx context.Context, pageID string) (*ConfluencePage, error)
	DeletePage(ctx context.Context, pageID string) error

	// Search
	SearchPages(ctx context.Context, cql string, limit int) ([]*ConfluencePage, error)

	// Content conversion
	ConvertToStorageFormat(ctx context.Context, content string, format string) (string, error)

	// Attachments
	AttachFile(ctx context.Context, pageID string, filename string, contentType string, data []byte) error
	GetAttachments(ctx context.Context, pageID string) ([]*ConfluenceAttachment, error)

	// Spaces
	GetSpaces(ctx context.Context) ([]*ConfluenceSpace, error)
	GetSpace(ctx context.Context, spaceKey string) (*ConfluenceSpace, error)

	// Labels
	AddLabels(ctx context.Context, pageID string, labels []string) error
	RemoveLabel(ctx context.Context, pageID string, label string) error
}

// ConfluencePage represents a Confluence page
type ConfluencePage struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title"`
	SpaceKey  string `json:"space_key"`
	ParentID  string `json:"parent_id,omitempty"`
	Body      string `json:"body"`         // In storage format
	Version   int    `json:"version,omitempty"`
	Status    string `json:"status,omitempty"` // current, draft
	URL       string `json:"url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ConfluencePageUpdate represents an update to a Confluence page
type ConfluencePageUpdate struct {
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	Version int    `json:"version"` // Required for optimistic locking
	Message string `json:"message,omitempty"` // Update message
}

// ConfluenceSpace represents a Confluence space
type ConfluenceSpace struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // global, personal
	HomepageID  string `json:"homepage_id,omitempty"`
}

// ConfluenceAttachment represents a file attachment
type ConfluenceAttachment struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	MediaType   string `json:"media_type"`
	FileSize    int64  `json:"file_size"`
	DownloadURL string `json:"download_url"`
}

// ConfluenceClientConfig represents configuration for Confluence client
type ConfluenceClientConfig struct {
	BaseURL string `yaml:"BASE_URL"`
	CloudID string `yaml:"CLOUD_ID"`
	// AccessToken contains OAuth credentials - do not log or expose
	AccessToken string // Set at runtime from OAuth
	Timeout     int    `yaml:"TIMEOUT_SECONDS"`
	MaxRetries  int    `yaml:"MAX_RETRIES"`
}

// ConfluenceService defines high-level Confluence integration operations
type ConfluenceService interface {
	// PRD export
	ExportPRDToPage(ctx context.Context, prdID string, spaceKey string, parentPageID string) (*ConfluencePage, error)

	// Update existing page from PRD
	UpdatePageFromPRD(ctx context.Context, prdID string, pageID string) (*ConfluencePage, error)

	// Template-based page creation
	CreatePageFromTemplate(ctx context.Context, templateID string, spaceKey string, variables map[string]interface{}) (*ConfluencePage, error)
}

// PRDToConfluenceConverter converts PRD content to Confluence storage format
type PRDToConfluenceConverter interface {
	// Convert converts PRD content to Confluence storage format (XHTML)
	Convert(prdContent interface{}) (string, error)

	// ConvertSection converts a single PRD section
	ConvertSection(sectionName string, content interface{}) (string, error)
}

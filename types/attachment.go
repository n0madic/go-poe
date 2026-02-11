package types

// Attachment included in a protocol message
type Attachment struct {
	URL           string  `json:"url"`
	ContentType   string  `json:"content_type"`
	Name          string  `json:"name"`
	InlineRef     *string `json:"inline_ref,omitempty"`
	ParsedContent *string `json:"parsed_content,omitempty"`
}

// AttachmentUploadResponse is the result of a file upload
type AttachmentUploadResponse struct {
	AttachmentURL *string `json:"attachment_url,omitempty"`
	MimeType      *string `json:"mime_type,omitempty"`
	Name          *string `json:"name,omitempty"`
	InlineRef     *string `json:"inline_ref,omitempty"`
}

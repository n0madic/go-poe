package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/n0madic/go-poe/types"
)

const defaultUploadBaseURL = "https://www.quora.com/poe_api/"

// UploadFileOptions configures file upload
type UploadFileOptions struct {
	File           io.Reader
	FileURL        string
	FileName       string
	APIKey         string
	NumTries       int
	RetrySleepTime time.Duration
	BaseURL        string
	ExtraHeaders   map[string]string
	HTTPClient     *http.Client
}

func (o *UploadFileOptions) defaults() {
	if o.NumTries <= 0 {
		o.NumTries = defaultNumTries
	}
	if o.RetrySleepTime <= 0 {
		o.RetrySleepTime = defaultRetrySleep
	}
	if o.BaseURL == "" {
		o.BaseURL = defaultUploadBaseURL
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
}

// UploadFile uploads a file to Poe and returns an Attachment
func UploadFile(ctx context.Context, opts *UploadFileOptions) (*types.Attachment, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("api_key is required (generate one at https://poe.com/api_key)")
	}
	if opts.File == nil && opts.FileURL == "" {
		return nil, fmt.Errorf("provide either File or FileURL")
	}
	if opts.File != nil && opts.FileURL != "" {
		return nil, fmt.Errorf("provide either File or FileURL, not both")
	}

	opts.defaults()
	endpoint := strings.TrimRight(opts.BaseURL, "/") + "/file_upload_3RD_PARTY_POST"

	var lastErr error
	for attempt := 0; attempt < opts.NumTries; attempt++ {
		att, err := doUpload(ctx, opts, endpoint)
		if err == nil {
			return att, nil
		}
		lastErr = err
		log.Printf("Upload attempt %d/%d failed: %v", attempt+1, opts.NumTries, err)
		if attempt < opts.NumTries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(opts.RetrySleepTime):
			}
		}
	}
	return nil, lastErr
}

func doUpload(ctx context.Context, opts *UploadFileOptions, endpoint string) (*types.Attachment, error) {
	var req *http.Request
	var err error

	if opts.FileURL != "" {
		// URL mode: POST form data
		form := strings.NewReader(fmt.Sprintf("download_url=%s&download_filename=%s", opts.FileURL, opts.FileName))
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint, form)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		// File mode: multipart upload
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("file", opts.FileName)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, opts.File); err != nil {
			return nil, err
		}
		writer.Close()

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}

	// Note: Authorization is raw key, NOT Bearer
	req.Header.Set("Authorization", opts.APIKey)
	for k, v := range opts.ExtraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return nil, &AttachmentUploadError{Message: fmt.Sprintf("HTTP error: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &AttachmentUploadError{
			Message: fmt.Sprintf("%d %s: %s", resp.StatusCode, resp.Status, string(body)),
		}
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &AttachmentUploadError{Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	attURL, _ := result["attachment_url"].(string)
	mimeType, _ := result["mime_type"].(string)

	if attURL == "" || mimeType == "" {
		return nil, &AttachmentUploadError{Message: fmt.Sprintf("unexpected response format: %v", result)}
	}

	name := opts.FileName
	if name == "" {
		name = "file"
	}

	return &types.Attachment{
		URL:         attURL,
		ContentType: mimeType,
		Name:        name,
	}, nil
}

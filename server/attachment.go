package server

import (
	"fmt"
	"strings"

	"github.com/n0madic/go-poe/types"
)

// InsertAttachmentMessages inserts messages containing attachment contents before the last user message.
func InsertAttachmentMessages(req *types.QueryRequest) *types.QueryRequest {
	if len(req.Query) == 0 {
		return req
	}

	lastMessage := req.Query[len(req.Query)-1]
	var textAttachmentMessages []types.ProtocolMessage
	var imageAttachmentMessages []types.ProtocolMessage

	for _, attachment := range lastMessage.Attachments {
		if attachment.ParsedContent == nil || *attachment.ParsedContent == "" {
			continue
		}
		parsedContent := *attachment.ParsedContent

		if attachment.ContentType == "text/html" {
			content := fmt.Sprintf(types.URLAttachmentTemplate, attachment.Name, parsedContent)
			textAttachmentMessages = append(textAttachmentMessages, types.ProtocolMessage{
				Role:    "user",
				Sender:  &types.Sender{},
				Content: content,
			})
		} else if strings.HasPrefix(attachment.ContentType, "text/") || attachment.ContentType == "application/pdf" {
			content := fmt.Sprintf(types.TextAttachmentTemplate, attachment.Name, parsedContent)
			textAttachmentMessages = append(textAttachmentMessages, types.ProtocolMessage{
				Role:    "user",
				Sender:  &types.Sender{},
				Content: content,
			})
		} else if strings.Contains(attachment.ContentType, "image") {
			var filename, description string
			parts := strings.SplitN(parsedContent, "***", 2)
			if len(parts) == 2 {
				filename = parts[0]
				description = parts[1]
			} else {
				filename = attachment.Name
				description = parsedContent
			}
			content := fmt.Sprintf(types.ImageVisionAttachmentTemplate, filename, description)
			imageAttachmentMessages = append(imageAttachmentMessages, types.ProtocolMessage{
				Role:    "user",
				Sender:  &types.Sender{},
				Content: content,
			})
		}
	}

	// Build new query: original messages (minus last) + text attachments + image attachments + last message
	newQuery := make([]types.ProtocolMessage, 0, len(req.Query)+len(textAttachmentMessages)+len(imageAttachmentMessages))
	newQuery = append(newQuery, req.Query[:len(req.Query)-1]...)
	newQuery = append(newQuery, textAttachmentMessages...)
	newQuery = append(newQuery, imageAttachmentMessages...)
	newQuery = append(newQuery, lastMessage)

	// Copy the request with the new query
	newReq := *req
	newReq.Query = newQuery
	return &newReq
}

// MakePromptAuthorRoleAlternated merges consecutive same-role messages.
func MakePromptAuthorRoleAlternated(messages []types.ProtocolMessage) []types.ProtocolMessage {
	var result []types.ProtocolMessage

	for _, msg := range messages {
		if len(result) > 0 && msg.Role == result[len(result)-1].Role {
			prev := result[len(result)-1]
			newContent := prev.Content + "\n\n" + msg.Content

			// Deduplicate attachments by URL
			addedURLs := make(map[string]bool)
			var newAttachments []types.Attachment
			for _, att := range msg.Attachments {
				if !addedURLs[att.URL] {
					addedURLs[att.URL] = true
					newAttachments = append(newAttachments, att)
				}
			}
			for _, att := range prev.Attachments {
				if !addedURLs[att.URL] {
					addedURLs[att.URL] = true
					newAttachments = append(newAttachments, att)
				}
			}

			prev.Content = newContent
			prev.Attachments = newAttachments
			result[len(result)-1] = prev
		} else {
			result = append(result, msg)
		}
	}

	return result
}

package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/n0madic/go-poe/sse"
	"github.com/n0madic/go-poe/types"
)

func handleQuery(ctx context.Context, w http.ResponseWriter, bot PoeBot, req *types.QueryRequest) {
	// Insert attachment messages if configured
	if bot.ShouldInsertAttachmentMessages() {
		req = InsertAttachmentMessages(req)
	}

	sseWriter := sse.NewWriter(w)

	// Get response channel from bot
	ch := bot.GetResponse(ctx, req)

	// Consume events and write SSE
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in bot response: %v", r)
				writeErrorEvent(sseWriter, "The bot encountered an unexpected issue.", false, nil)
			}
		}()

		for event := range ch {
			switch e := event.(type) {
			case *types.PartialResponse:
				// If there's an attachment, emit file event first
				if e.Attachment != nil {
					writeFileEvent(sseWriter, e.Attachment)
				}

				if e.IsSuggestedReply {
					writeSuggestedReplyEvent(sseWriter, e.Text)
				} else if e.IsReplaceResponse {
					writeReplaceResponseEvent(sseWriter, e.Text)
				} else {
					writeTextEvent(sseWriter, e.Text, e.Index)
				}

			case *types.ErrorResponse:
				writeErrorEvent(sseWriter, e.Text, e.AllowRetry, e.ErrorType)

			case *types.MetaResponse:
				writeMetaEvent(sseWriter, e)

			case *types.DataResponse:
				writeDataEvent(sseWriter, e.Metadata)
			}
		}
	}()

	// Always emit done event
	writeDoneEvent(sseWriter)
}

func writeTextEvent(w *sse.Writer, text string, index *int) {
	data := map[string]any{"text": text}
	if index != nil {
		data["index"] = *index
	}
	b, _ := json.Marshal(data)
	w.WriteEvent(sse.Event{Event: "text", Data: string(b)})
}

func writeReplaceResponseEvent(w *sse.Writer, text string) {
	b, _ := json.Marshal(map[string]any{"text": text})
	w.WriteEvent(sse.Event{Event: "replace_response", Data: string(b)})
}

func writeSuggestedReplyEvent(w *sse.Writer, text string) {
	b, _ := json.Marshal(map[string]any{"text": text})
	w.WriteEvent(sse.Event{Event: "suggested_reply", Data: string(b)})
}

func writeFileEvent(w *sse.Writer, att *types.Attachment) {
	data := map[string]any{
		"url":          att.URL,
		"content_type": att.ContentType,
		"name":         att.Name,
	}
	if att.InlineRef != nil {
		data["inline_ref"] = *att.InlineRef
	}
	b, _ := json.Marshal(data)
	w.WriteEvent(sse.Event{Event: "file", Data: string(b)})
}

func writeMetaEvent(w *sse.Writer, meta *types.MetaResponse) {
	b, _ := json.Marshal(map[string]any{
		"content_type":      meta.ContentType,
		"refetch_settings":  meta.RefetchSettings,
		"linkify":           meta.Linkify,
		"suggested_replies": meta.SuggestedReplies,
	})
	w.WriteEvent(sse.Event{Event: "meta", Data: string(b)})
}

func writeDataEvent(w *sse.Writer, metadata string) {
	b, _ := json.Marshal(map[string]any{"metadata": metadata})
	w.WriteEvent(sse.Event{Event: "data", Data: string(b)})
}

func writeErrorEvent(w *sse.Writer, text string, allowRetry bool, errorType *string) {
	data := map[string]any{"allow_retry": allowRetry}
	if text != "" {
		data["text"] = text
	}
	if errorType != nil {
		data["error_type"] = *errorType
	}
	b, _ := json.Marshal(data)
	w.WriteEvent(sse.Event{Event: "error", Data: string(b)})
}

func writeDoneEvent(w *sse.Writer) {
	w.WriteEvent(sse.Event{Event: "done", Data: "{}"})
}

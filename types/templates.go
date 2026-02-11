package types

// Templates for formatting attachment content into LLM-consumable messages.
// Use with fmt.Sprintf.
const (
	TextAttachmentTemplate = "Below is the content of %s:\n\n%s"

	URLAttachmentTemplate = "Assume you can access the external URL %s. " +
		"Use the URL's content below to respond to the queries:\n\n%s"

	ImageVisionAttachmentTemplate = "I have uploaded an image (%s). " +
		"Assume that you can see the attached image. " +
		"First, read the image analysis:\n\n" +
		"<image_analysis>%s</image_analysis>\n\n" +
		"Use any relevant parts to inform your response. " +
		"Do NOT reference the image analysis in your response. " +
		"Respond in the same language as my next message. "
)

package messenger

import (
	"fmt"
	"strings"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FileAttachmentData represents file attachment data
type FileAttachmentData struct {
	ID       int64
	Filename string
	MimeType string
	FileSize int64
	URL      string
}

// FileAttachment renders a file attachment
func FileAttachment(r *ui.Request, attachment FileAttachmentData) Node {
	// Determine file type
	isImage := strings.HasPrefix(attachment.MimeType, "image/")
	isPDF := attachment.MimeType == "application/pdf"
	isText := strings.HasPrefix(attachment.MimeType, "text/")

	// Format file size
	sizeStr := formatFileSize(attachment.FileSize)

	if isImage {
		return imageAttachment(r, attachment, sizeStr)
	} else if isPDF {
		return pdfAttachment(r, attachment, sizeStr)
	} else if isText {
		return textAttachment(r, attachment, sizeStr)
	} else {
		return genericAttachment(r, attachment, sizeStr)
	}
}

// imageAttachment renders an image attachment with preview
func imageAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 rounded-lg overflow-hidden border border-base-300"),
		Div(
			Class("relative"),
			// Image preview
			Img(
				Src(attachment.URL),
				Alt(attachment.Filename),
				Class("max-w-full h-auto cursor-pointer"),
				Attr("onclick", fmt.Sprintf("window.open('%s', '_blank')", attachment.URL)),
				Title("Click to view full size"),
			),
		),
		Div(
			Class("p-2 bg-base-200 flex items-center justify-between text-xs"),
			Span(
				Class("text-base-content/70 truncate"),
				Text(attachment.Filename),
			),
			Span(
				Class("text-base-content/50 ml-2"),
				Text(sizeStr),
			),
		),
	)
}

// pdfAttachment renders a PDF attachment
func pdfAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"),
		Div(
			Class("text-3xl"),
			Text("📄"),
		),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(fmt.Sprintf("PDF • %s", sizeStr)),
			),
		),
		A(
			Href(attachment.URL),
			Target("_blank"),
			Class("btn btn-sm btn-primary"),
			Text("Open"),
		),
	)
}

// textAttachment renders a text file attachment
func textAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"),
		Div(
			Class("text-3xl"),
			Text("📝"),
		),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(fmt.Sprintf("Text • %s", sizeStr)),
			),
		),
		A(
			Href(attachment.URL),
			Target("_blank"),
			Class("btn btn-sm btn-primary"),
			Text("View"),
		),
	)
}

// genericAttachment renders a generic file attachment
func genericAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	// Get file extension for icon
	ext := getFileExtension(attachment.Filename)
	icon := getFileIcon(ext)

	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"),
		Div(
			Class("text-3xl"),
			Text(icon),
		),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(fmt.Sprintf("%s • %s", strings.ToUpper(ext), sizeStr)),
			),
		),
		A(
			Href(attachment.URL),
			Target("_blank"),
			Class("btn btn-sm btn-primary"),
			Text("Download"),
		),
	)
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getFileExtension extracts file extension from filename
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// getFileIcon returns an emoji icon based on file extension
func getFileIcon(ext string) string {
	ext = strings.ToLower(ext)
	icons := map[string]string{
		"doc":  "📄",
		"docx": "📄",
		"xls":  "📊",
		"xlsx": "📊",
		"ppt":  "📽️",
		"pptx": "📽️",
		"zip":  "📦",
		"rar":  "📦",
		"7z":   "📦",
		"mp3":  "🎵",
		"mp4":  "🎬",
		"avi":  "🎬",
		"mov":  "🎬",
		"exe":  "⚙️",
		"dmg":  "💿",
		"iso":  "💿",
	}
	if icon, ok := icons[ext]; ok {
		return icon
	}
	return "📎" // Default file icon
}

package domain

import "time"

type Attachment struct {
	ID           string     `json:"id"`
	CardID       string     `json:"card_id"`
	UploadedBy   string     `json:"uploaded_by"`
	Filename     string     `json:"filename"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	FileSize     int64      `json:"file_size"`
	ObjectKey    string     `json:"-"`
	URL          string     `json:"url"`
	IsCover      bool       `json:"is_cover"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"-"`

	Uploader *User `json:"uploader,omitempty"`
}

var AllowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"text/plain":      true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

var ImageMimeTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

const MaxFileSize = int64(10 * 1024 * 1024) // 10MB

func IsImageMimeType(mimeType string) bool {
	for _, t := range ImageMimeTypes {
		if t == mimeType {
			return true
		}
	}
	return false
}

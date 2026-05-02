package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/codewebkhongkho/trello-agent/internal/domain"
	"github.com/codewebkhongkho/trello-agent/internal/repository"
	"github.com/codewebkhongkho/trello-agent/pkg/apperror"
	"github.com/codewebkhongkho/trello-agent/pkg/cuid"
	"github.com/codewebkhongkho/trello-agent/pkg/storage"
)

type AttachmentService struct {
	attachmentRepo *repository.AttachmentRepository
	cardRepo       *repository.CardRepository
	boardRepo      repository.BoardRepository
	activityRepo   *repository.ActivityRepository
	storage        storage.StorageService
}

func NewAttachmentService(
	attachmentRepo *repository.AttachmentRepository,
	cardRepo *repository.CardRepository,
	boardRepo repository.BoardRepository,
	activityRepo *repository.ActivityRepository,
	storage storage.StorageService,
) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		cardRepo:       cardRepo,
		boardRepo:      boardRepo,
		activityRepo:   activityRepo,
		storage:        storage,
	}
}

func (s *AttachmentService) ListByCard(ctx context.Context, userID, cardID string) ([]*domain.Attachment, error) {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, apperror.ErrForbidden
	}

	return s.attachmentRepo.FindByCardID(ctx, cardID)
}

func (s *AttachmentService) Upload(ctx context.Context, userID, cardID string, file *multipart.FileHeader) (*domain.Attachment, error) {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess || !role.CanEdit() {
		return nil, apperror.ErrForbidden
	}

	// Validate file size
	if file.Size > domain.MaxFileSize {
		return nil, apperror.New("FILE_TOO_LARGE", "File too large (max 10MB)", 400)
	}

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if !domain.AllowedMimeTypes[mimeType] {
		return nil, apperror.New("INVALID_FILE_TYPE", "File type not allowed", 400)
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", cuid.New(), ext)
	objectKey := fmt.Sprintf("attachments/%s/%s/%s", boardID, cardID, filename)

	// Upload to storage
	url, err := s.storage.Upload(ctx, objectKey, src, file.Size, mimeType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	attachment := &domain.Attachment{
		ID:           cuid.New(),
		CardID:       cardID,
		UploadedBy:   userID,
		Filename:     filename,
		OriginalName: file.Filename,
		MimeType:     mimeType,
		FileSize:     file.Size,
		ObjectKey:    objectKey,
		URL:          url,
		IsCover:      false,
		CreatedAt:    now,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		return nil, err
	}

	s.logActivity(ctx, boardID, &cardID, userID, domain.ActivityAttachmentAdded, map[string]interface{}{
		"attachment_id":   attachment.ID,
		"attachment_name": attachment.OriginalName,
	})

	return s.attachmentRepo.FindByID(ctx, attachment.ID)
}

func (s *AttachmentService) Delete(ctx context.Context, userID, attachmentID string) error {
	attachment, err := s.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, attachment.CardID)
	if err != nil {
		return err
	}

	// Check if user is uploader or board admin
	isUploader := attachment.UploadedBy == userID
	_, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	isAdmin := role == domain.BoardRoleAdmin || role == domain.BoardRoleOwner

	if !isUploader && !isAdmin {
		return apperror.ErrForbidden
	}

	if err := s.attachmentRepo.SoftDelete(ctx, attachmentID); err != nil {
		return err
	}

	s.logActivity(ctx, boardID, &attachment.CardID, userID, domain.ActivityAttachmentDeleted, map[string]interface{}{
		"attachment_name": attachment.OriginalName,
	})

	return nil
}

func (s *AttachmentService) SetCover(ctx context.Context, userID, attachmentID string) error {
	attachment, err := s.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return err
	}

	// Validate it's an image
	if !domain.IsImageMimeType(attachment.MimeType) {
		return apperror.New("INVALID_COVER_TYPE", "Only images can be set as cover", 400)
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, attachment.CardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	if err := s.attachmentRepo.SetCover(ctx, attachment.CardID, attachmentID); err != nil {
		return err
	}

	s.logActivity(ctx, boardID, &attachment.CardID, userID, domain.ActivityCoverSet, map[string]interface{}{
		"attachment_id": attachmentID,
	})

	return nil
}

func (s *AttachmentService) RemoveCover(ctx context.Context, userID, cardID string) error {
	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, cardID)
	if err != nil {
		return err
	}

	canAccess, role, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !canAccess || !role.CanEdit() {
		return apperror.ErrForbidden
	}

	if err := s.attachmentRepo.RemoveCover(ctx, cardID); err != nil {
		return err
	}

	s.logActivity(ctx, boardID, &cardID, userID, domain.ActivityCoverRemoved, nil)

	return nil
}

func (s *AttachmentService) GetDownloadURL(ctx context.Context, userID, attachmentID string) (string, error) {
	attachment, err := s.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return "", err
	}

	boardID, err := s.cardRepo.FindBoardIDByCardID(ctx, attachment.CardID)
	if err != nil {
		return "", err
	}

	canAccess, _, err := s.boardRepo.CanUserAccess(ctx, boardID, userID)
	if err != nil {
		return "", err
	}
	if !canAccess {
		return "", apperror.ErrForbidden
	}

	return attachment.URL, nil
}

func (s *AttachmentService) logActivity(ctx context.Context, boardID string, cardID *string, userID string, action domain.ActivityAction, metadata map[string]interface{}) {
	activity := &domain.ActivityLog{
		ID:        cuid.New(),
		BoardID:   boardID,
		CardID:    cardID,
		UserID:    userID,
		Action:    action,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	s.activityRepo.Create(ctx, activity)
}

func sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	return filename
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

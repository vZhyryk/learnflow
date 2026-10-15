package contentrepository

import (
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanContentItem(row repository.RowScanner) (*contentdomain.ContentItem, error) {
	content := &contentdomain.ContentItem{}
	err := row.Scan(
		&content.ID,
		&content.Slug,
		&content.Title,
		&content.ContentType,
		&content.Description,
		&content.Body,
		&content.VideoURL,
		&content.FileURL,
		&content.EstimatedMinutes,
		&content.EstimatedPages,
		&content.ThumbnailURL,
		&content.SeoTitle,
		&content.SeoDescription,
		&content.OgImageURL,
		&content.CanonicalURL,
		&content.IsIndexable,
		&content.Status,
		&content.CreatedByUserID,
		&content.CreatedAt,
		&content.UpdatedAt,
		&content.UpdatedByUserID,
		&content.PublishedAt,
		&content.PublishedByUserID,
		&content.DeletedAt,
		&content.DeletedByUserID,
		&content.ArchivedAt,
		&content.ArchivedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return content, nil
}

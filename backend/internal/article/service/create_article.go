package articleservice

import (
	"context"
	"errors"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
)

// CreateArticle creates a new draft article after checking the slug is not already in use.
func (s *Service) CreateArticle(ctx context.Context, req articledomain.CreateArticleRequest) (string, error) {
	var articleID string
	err := s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		existing, err := s.articleRepo.GetArticleBySlug(ctx, req.Slug)
		if err != nil && !errors.Is(err, articledomain.ErrArticleNotFound) {
			return fmt.Errorf("service.CreateArticle: %w", err)
		}
		if existing != nil {
			return articledomain.ErrInvalidSlug
		}

		isIndexable := true
		if req.IsIndexable != nil {
			isIndexable = *req.IsIndexable
		}

		article := articledomain.Article{
			Slug:            req.Slug,
			Title:           req.Title,
			Description:     req.Description,
			SeoTitle:        req.SeoTitle,
			SeoDescription:  req.SeoDescription,
			OgImageURL:      req.OgImageURL,
			IsIndexable:     isIndexable,
			CreatedByUserID: req.CreatedByUserID,
			Body:            req.Body,
		}

		createdArticle, err := s.articleRepo.CreateArticle(ctx, &article)
		if err != nil {
			return fmt.Errorf("service.CreateArticle: %w", err)
		}

		articleID = createdArticle.ID
		return s.actionRepo.CreateAdminAction(ctx, &auditdomain.AdminAction{
			AdminUserID: req.CreatedByUserID,
			ActionType:  auditdomain.ActionCreateItem,
			TargetType:  auditdomain.TargetArticle,
			TargetID:    articleID,
		})
	})

	return articleID, err
}

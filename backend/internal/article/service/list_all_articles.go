package articleservice

import (
	"context"
	"fmt"
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/pagination"
)

// GetAllArticles returns article filtered by status (archived, published, draft, or all).
func (s *Service) GetAllArticles(ctx context.Context, getType articledomain.ArticleStatus, params pagination.Params) (articleList []*articledomain.Article, err error) {
	switch getType {
	case articledomain.ArchivedStatus:
		articleList, err = s.articleRepo.GetAllArchivedArticles(ctx, params)
	case articledomain.PublishedStatus:
		articleList, err = s.articleRepo.GetAllPublishedArticles(ctx, params)
	case articledomain.DraftStatus:
		articleList, err = s.articleRepo.GetAllDraftArticles(ctx, params)
	default:
		articleList, err = s.articleRepo.GetAllArticles(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("service.GetAllArticles: %w", err)
	}
	return articleList, nil
}

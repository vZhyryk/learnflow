package articlerepository

import (
	articledomain "learnflow_backend/internal/article/domain"
	"learnflow_backend/internal/shared/repository"
)

func scanArticle(row repository.RowScanner) (*articledomain.Article, error) {
	article := &articledomain.Article{}
	err := row.Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.Excerpt,
		&article.Body,
		&article.SeoTitle,
		&article.SeoDescription,
		&article.OgImageURL,
		&article.IsIndexable,
		&article.Status,
		&article.CreatedByUserID,
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.PublishedAt,
		&article.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return article, nil
}

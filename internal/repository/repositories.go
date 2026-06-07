package repository

import (
	"context"
	"infolinks-backend/internal/models"
)

type ContentRepository interface {
	Get(ctx context.Context) ([]byte, error)
}

type SEORepository interface {
	GetCoursePageByCode(ctx context.Context, code string) (*CoursePageData, error)
	ListCourseCodesForSitemap(ctx context.Context) ([]string, error)
	ListProgramsForSitemap(ctx context.Context, slugFn func(string) string) ([]ProgramSitemapEntry, error)
	ListCoursesIndex(ctx context.Context) ([]CourseIndexEntry, error)
	GetProgramBySlug(ctx context.Context, slug string, slugFn func(string) string) (*ProgramPageData, error)
}

type LinkClickRepository interface {
	List(ctx context.Context) ([]models.LinkClick, error)
	Create(ctx context.Context, lc models.LinkClick) error
}

type PageViewRepository interface {
	List(ctx context.Context) ([]models.PageView, error)
	Create(ctx context.Context, pv models.PageView) error
}

type LinkRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, link models.Link) error
	Update(ctx context.Context, link models.Link, id int) error
}

type ExtraSectionRepository interface {
	List(ctx context.Context) ([]models.ExtraSection, error)
	Create(ctx context.Context, section models.ExtraSection) error
	Update(ctx context.Context, section models.ExtraSection, id int) error
	Delete(ctx context.Context, id int) error
}

type ExtraLinkRepository interface {
	List(ctx context.Context) ([]models.ExtraLink, error)
	Create(ctx context.Context, link models.ExtraLink) error
	Update(ctx context.Context, link models.ExtraLink, id int) error
	Delete(ctx context.Context, id int) error
}

type CourseRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, course models.Course) error
	GetByID(ctx context.Context, id int) (models.Course, error)
	Update(ctx context.Context, course models.Course, id int) error
}

type ReportRepository interface {
	Delete(ctx context.Context, id int) error
	Create(ctx context.Context, report models.Report) error
	Update(ctx context.Context, status string, id int) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Report, error)
}

type ContributionRepository interface {
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, status string, id int) error
	Create(ctx context.Context, contribution models.Contribution) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Contribution, error)
}

type FeedbackRepository interface {
	Delete(ctx context.Context, id int) error
	Update(ctx context.Context, status string, id int) error
	Create(ctx context.Context, feedback models.Feedback) error
	List(ctx context.Context, limit int, offset int, q string, status string) ([]models.Feedback, error)
}

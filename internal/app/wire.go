package app

import (
	"database/sql"

	"infolinks-backend/internal/api"
	"infolinks-backend/internal/repository"
	"infolinks-backend/internal/service"
)

// Wire builds repository-backed services for production and integration tests.
func Wire(db *sql.DB) (api.Dependencies, *service.UserService) {
	userRepo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo)

	analyticsRepo := repository.NewPostgresAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo)

	linkRepo := repository.NewPostgresLinkRepository(db)
	linkService := service.NewLinkService(linkRepo)

	courseRepo := repository.NewPostgresCourseRepository(db)
	courseService := service.NewCourseService(courseRepo)

	reportRepo := repository.NewPostgresReportRepository(db)
	reportService := service.NewReportService(reportRepo)

	feedbackRepo := repository.NewPostgresFeedbackRepository(db)
	feedbackService := service.NewFeedbackService(feedbackRepo)

	contentRepo := repository.NewPostgresContentRepository(db)
	contentService := service.NewContentService(contentRepo)

	pageViewRepo := repository.NewPostgresPageViewRepository(db)
	pageViewService := service.NewPageViewService(pageViewRepo)

	linkClickRepo := repository.NewPostgresLinkClickRepository(db)
	linkClickService := service.NewLinkClickService(linkClickRepo)

	contributionsRepo := repository.NewPostgresContributionRepository(db)
	contributionsService := service.NewContributionService(contributionsRepo)

	extraSectionRepo := repository.NewPostgresExtraSectionRepository(db)
	extraSectionService := service.NewExtraSectionService(extraSectionRepo)

	extraLinkRepo := repository.NewPostgresExtraLinkRepository(db)
	extraLinkService := service.NewExtraLinkService(extraLinkRepo)

	serviceRepo := repository.NewPostgresServiceRepository(db)
	serviceService := service.NewServiceService(serviceRepo)

	return api.Dependencies{
		UserService:         userService,
		AnalyticsService:    analyticsService,
		LinkService:         linkService,
		CourseService:       courseService,
		ReportService:       reportService,
		ContentService:      contentService,
		FeedbackService:     feedbackService,
		PageViewService:     pageViewService,
		LinkClickService:    linkClickService,
		ContributionService: contributionsService,
		ExtraSectionService: extraSectionService,
		ExtraLinkService:    extraLinkService,
		ServiceService:      serviceService,
	}, userService
}

package errs

import "errors"

var (
	// Reports Errors
	ErrReportNotFound      = errors.New("report not found")
	ErrInvalidReportID     = errors.New("invalid report id")
	ErrInvalidReportStatus = errors.New("status must be open or resolved")

	// Contributions Errors
	ErrContributionNotFound      = errors.New("contribution not found")
	ErrInvalidContributionID     = errors.New("invalid contribution id")
	ErrInvalidContributionStatus = errors.New("status must be pending or approved")

	// Common Errors
	ErrDatabaseDown                 = errors.New("db is down")
	ErrStatusRequired               = errors.New("status is required")
	ErrCourseNameAndLinkUrlRequired = errors.New("course_name and link_url are required")
	ErrInvalidParams                = errors.New("limit should be between 1-100 and offset >= 0")
)

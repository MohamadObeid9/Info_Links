package errs

import "errors"

var (
	// Reports Errors
	ErrReportNotFound      = errors.New("report not found")
	ErrReportInvalidID     = errors.New("invalid report id")
	ErrReportInvalidStatus = errors.New("status must be open or resolved")

	// Contributions Errors
	ErrContributionNotFound      = errors.New("contribution not found")
	ErrContributionInvalidID     = errors.New("invalid contribution id")
	ErrContributionInvalidStatus = errors.New("status must be pending or approved")

	// Feedback Errors
	ErrFeedbackNotFound                  = errors.New("feedback not found")
	ErrFeedbackInvalidID                 = errors.New("invalid feedback id")
	ErrFeedbackInvalidStatus             = errors.New("status must be new or read")
	ErrFeedbackInvalidRating             = errors.New("rating should be between 1 and 5")
	ErrFeedbackCategoryAndRatingRequired = errors.New("category and rating are required")
	ErrFeedbackInvalidCategory           = errors.New("category must be one of the following : ui/ux or content or functionality or performance or accessibility")

	// Course Errors
	ErrCourseNotFound            = errors.New("course not found")
	ErrCourseInvalidID           = errors.New("course invalid id")
	ErrCourseCodeAndNameRequired = errors.New("course code and course name are required")

	// Common Errors
	ErrDatabaseDown                 = errors.New("db is down")
	ErrStatusRequired               = errors.New("status is required")
	ErrCourseNameAndLinkUrlRequired = errors.New("course_name and link_url are required")
	ErrInvalidParams                = errors.New("limit should be between 1-100 and offset >= 0")
)

package errs

import "errors"

// Link clicks errors
var (
	ErrLinkClickLinkIDAndExtraLinkIDRequired = errors.New("link id and extra link id are required")
	ErrLinkClickLinkIDAndExtraLinkIDSet      = errors.New("link id and extra link id cannot be set at the same time")
)

// Reports Errors
var (
	ErrReportNotFound      = errors.New("report not found")
	ErrReportInvalidID     = errors.New("invalid report id")
	ErrReportInvalidStatus = errors.New("status must be open or resolved")
)

// Links Errors
var (
	ErrLinkNotFound            = errors.New("link not found")
	ErrLinkInvalidID           = errors.New("invalid link id")
	ErrLinkURLAndLabelRequired = errors.New("link url and link label are required")
)

// Extra section errors
var (
	ErrExtraSectionNotFound      = errors.New("extra section not found")
	ErrExtraSectionInvalidID     = errors.New("invalid extra section id")
	ErrExtraSectionTitleRequired = errors.New("extra section title is required")
)

// Extra link errors
var (
	ErrExtraLinkNotFound            = errors.New("extra link not found")
	ErrExtraLinkInvalidID           = errors.New("invalid extra link id")
	ErrExtraLinkURLAndLabelRequired = errors.New("extra link url and label are required")
	ErrExtraLinkInvalidSectionID    = errors.New("extra link invalid section id")
)

// Course Errors
var (
	ErrCourseNotFound            = errors.New("course not found")
	ErrCourseInvalidID           = errors.New("course invalid id")
	ErrCourseInvalidSemestreID   = errors.New("course invalid semestre id")
	ErrCoursePatchEmpty          = errors.New("course invalid update parameters")
	ErrCourseCodeAndNameRequired = errors.New("course code and course name are required")
)

// Contributions Errors
var (
	ErrContributionNotFound      = errors.New("contribution not found")
	ErrContributionInvalidID     = errors.New("invalid contribution id")
	ErrContributionInvalidStatus = errors.New("status must be pending or approved")
)

// Common Errors
var (
	ErrDatabaseDown                 = errors.New("db is down")
	ErrStatusRequired               = errors.New("status is required")
	ErrCourseNameAndLinkUrlRequired = errors.New("course_name and link_url are required")
	ErrInvalidParams                = errors.New("limit should be between 1-100 and offset >= 0")
)

// Feedback Errors
var (
	ErrFeedbackNotFound                  = errors.New("feedback not found")
	ErrFeedbackInvalidID                 = errors.New("invalid feedback id")
	ErrFeedbackInvalidStatus             = errors.New("status must be new or read")
	ErrFeedbackInvalidRating             = errors.New("rating should be between 1 and 5")
	ErrFeedbackCategoryAndRatingRequired = errors.New("category and rating are required")
	ErrFeedbackInvalidCategory           = errors.New("category must be one of the following : ui/ux or content or functionality or performance or accessibility")
)

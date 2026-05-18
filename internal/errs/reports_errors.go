package errs

import "errors"

var (
	ErrDatabaseDown                    = errors.New("db is down")
	ErrInvalidReportID                 = errors.New("invalid report id")
	ErrReportNotFound                  = errors.New("report not found")
	ErrReportStatusRequired            = errors.New("status is required")
	ErrInvalidReportStatus             = errors.New("status must be open or resolved")
	ErrCourseNameAndLinkUrlAreRequired = errors.New("course_name and link_url are required")
	ErrListReportInvalidParams         = errors.New("limit should be between 1-100 and offset >= 0")
)

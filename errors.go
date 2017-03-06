package tigerblood

// errnos for APP-MOZLOG

const (
	_ = iota  // must be > 0 to indicate error https://wiki.mozilla.org/Firefox/Services/Logging#Application_Request_Summary_.28Type:.22request.summary.22.29
	HawkAuthFormatError
	HawkReplayError
	HawkErrNoAuth
	HawkInvalidHash
	HawkCredError

	HawkOtherAuthError

	HawkMissingContentType

	HawkValidationError

	HawkInvalidBodyHash
	HawkReadBodyError
)

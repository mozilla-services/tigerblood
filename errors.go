package tigerblood

// errno and descriptions for APP-MOZLOG
// must be > 0 to indicate error https://wiki.mozilla.org/Firefox/Services/Logging#Application_Request_Summary_.28Type:.22request.summary.22.29

type Errno uint

const (
	// auth errors
	HawkAuthFormatError = 1
	HawkReplayError = 2
	HawkErrNoAuth = 3
	HawkInvalidHash = 4
	HawkCredError = 5
	HawkOtherAuthError = 6
	HawkMissingContentType = 7
	HawkValidationError = 8
	HawkInvalidBodyHash = 9
	HawkReadBodyError = 10

	// context middleware errors
	RequestContextMissingDB = 20
	RequestContextMissingStatsd = 21
	RequestContextMissingViolations = 22

	// encoding/decoding errors
	BodyReadError = 30
	JSONMarshalError = 31
	JSONUnmarshalError = 32

	// validation errors
	InvalidIPError = 40
	InvalidReputationError = 41
	InvalidViolationTypeError = 42

	// missing parameter errors
	MissingIPError = 50
	MissingReputationError = 51
	MissingViolationTypeError = 52

	// IO/DB errors
	DBError = 60
	CWDNotFound = 61
	FileNotFound = 62
)


// Returns a format string for the errno
// not implemented for all errnos
func DescribeErrno(errno Errno) string {
	switch errno {

	case BodyReadError:
		return "Error reading the request body: %s"
	case JSONMarshalError:
		return "Error marshaling %s to JSON: %s"
	case JSONUnmarshalError:
		return "Error unmarshaling request body from JSON: %s"

	case InvalidIPError:
		return "Invalid IP: %s"
	case InvalidReputationError:
		return "Invalid reputation: %s"
	case InvalidViolationTypeError:
		return "Invalid violation type: %s"

	case MissingIPError:
		return "Error finding IP parameter in %s: %s"
	case MissingReputationError:
		return "Error finding reputation parameter in %s: %s"
	case MissingViolationTypeError:
		return "Error finding violation type in %s: %s"

	case RequestContextMissingDB:
		return "Could not find database handler in request context."
	case RequestContextMissingViolations:
		return "Could not find violation penalties in request context."
	case RequestContextMissingStatsd:
		return "Could not find statsdClient in request context."

	case CWDNotFound:
		return "Error getting CWD: %s"
	case FileNotFound:
		return "Error finding file %s: %s"

	default:
		return "Error: %s"
	}
}

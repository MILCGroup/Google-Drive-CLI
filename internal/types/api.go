package types

// RequestType classifies API requests for proper parameter injection
type RequestType string

const (
	RequestTypeGetByID          RequestType = "GetById"
	RequestTypeListOrSearch     RequestType = "ListOrSearch"
	RequestTypeMutation         RequestType = "Mutation"
	RequestTypeRevisionOp       RequestType = "RevisionOp"
	RequestTypePermissionOp     RequestType = "PermissionOp"
	RequestTypeDownloadOrExport RequestType = "DownloadOrExport"
	RequestTypeBatchOp          RequestType = "BatchOp"
)

// RequestContext carries context for API request shaping
type RequestContext struct {
	Profile           string
	DriveID           string
	InvolvedFileIDs   []string
	InvolvedParentIDs []string
	RequestType       RequestType
	TraceID           string
}

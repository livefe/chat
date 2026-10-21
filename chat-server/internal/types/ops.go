package types

type AdminPendingCountsResponse struct {
	IMUnread         int64 `json:"imUnread"`
	SupportQueue     int64 `json:"supportQueue"`
	ReportPending    int64 `json:"reportPending"`
	FeedbackPending  int64 `json:"feedbackPending"`
	RoomAuditPending int64 `json:"roomAuditPending"`
	PostAuditPending int64 `json:"postAuditPending"`
}

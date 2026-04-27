package form

type AdminRecipients struct {
	ID           int64   `form:"id" json:"id"`
	AdminID      int64   `form:"admin_id" json:"admin_id"`
	MediaID      int64   `form:"media_id" json:"media_id"`
	GroupID      int64   `form:"group_id" json:"group_id"`
	RecipientID  string  `form:"recipient_id" json:"recipient_id"`
	Interval     int     `form:"interval"`      // interval
	IntervalType string  `form:"interval_type"` // interval_type
	RelatedIDs   []int64 `form:"related_ids" json:"related_ids"`
	Mark         string  `form:"mark" json:"mark"`
	Status       bool    `form:"status" json:"status"`
}

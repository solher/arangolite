package requests

import "fmt"

// FollowCursor queries the next page of result for the given cursor.
type FollowCursor struct {
	Cursor string
}

func (r *FollowCursor) Path() string {
	return fmt.Sprintf("/_api/cursor/%s", r.Cursor)
}

func (r *FollowCursor) Method() string {
	return "PUT"
}

func (r *FollowCursor) Generate() []byte {
	return nil
}

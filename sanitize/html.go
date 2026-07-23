package sanitize

import "github.com/microcosm-cc/bluemonday"

var ugc = bluemonday.UGCPolicy()

// HTML returns content sanitized with bluemonday's UGC policy.
func HTML(s string) string {
	return ugc.Sanitize(s)
}

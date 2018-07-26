package label

import (
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{}

	log = env.NewLogger("label")
)

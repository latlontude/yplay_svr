package ddactivity

import (
	//"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
	//"/loadpage": auth.Apify5(doLoadPage),
	}

	log = env.NewLogger("ddactivity")
)

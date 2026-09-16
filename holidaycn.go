package holidaycn

import (
	"time"
)

var cnLocation *time.Location

func init() {
	var err error
	cnLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// Fallback to UTC+8 if timezone data is not available
		cnLocation = time.FixedZone("CST", 8*60*60)
	}
}

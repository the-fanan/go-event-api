package constants

import (
	"time"
)
const JWT_ACCESS_TOKEN_LIFETIME = time.Hour * 24 * 100
const JWT_REFRESH_TOKEN_LIFETIME = time.Hour * 24 * 365
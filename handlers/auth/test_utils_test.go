package auth

import (
	"github.com/arran4/goa4web/internal/sign"
)

func SignBackURL(key, u string, ts int64) string {
	return sign.Sign("back:"+u, key, sign.WithOutNonce())
}

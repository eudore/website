package middleware

import (
	"github.com/eudore/eudore"
	"github.com/eudore/website/util/uuid"
)

func NewAddHeaderFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		requestId := uuid.GetId()
		ctx.Request().Header.Add(eudore.HeaderXRequestID, requestId)
		// add default header
		h := ctx.Response().Header()
		h.Add(eudore.HeaderXRequestID, requestId)
		h.Add("Cache-Control", "no-cache")
		h.Add("X-XSS-Protection", "1; mode=block")
		h.Add("X-Frame-Options", "SAMEORIGIN")
		h.Add("X-Content-Type-Options", "nosniff")
		h.Add("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; report-uri /eudore/csp")
		h.Add("Referrer-Policy", "origin-when-cross-origin,strict-origin-when-cross-origin")
		/*	if ctx.Istls() {
			// h.Add("Strict-Transport-Security", "max-age=31536000;includesubdomains;preload")
			// h.Add("Public-Key-Pins", `pin-sha256="sMU3CCjru4a49HAhlUSFaR1ryqFCVzv/eScJ9sE8jqY="; pin-sha256="1WDPq2eHdQ+RNNmbZCKIxy/0POuXu8Vbd6OfCy1N6aA="; max-age=2592000; includeSubDomains; report-uri="https://www.wejass.com"`)
		}*/
	}
}

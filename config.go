package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func storeOptions() {
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,                    // Безопасность: cookie будет доступен только через HTTP
		Secure:   false,                   // Включить, если используется HTTPS
		SameSite: http.SameSiteStrictMode, // Защита от CSRF атак
	}
}

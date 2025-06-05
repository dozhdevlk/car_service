package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router) {
	// HTML страницы
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/partner-register.html", pagePartnerRegisterHandler)
	r.HandleFunc("/admin", pageAdminHandler)
	r.HandleFunc("/partner/{id}", pagePartnerHandler)
	r.HandleFunc("/dashboard/{id}", dashboardHandler)
	r.HandleFunc("/client", clientHandler)

	// API endpoints
	r.HandleFunc("/api/register", registerHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/api/user", userHandler).Methods("GET")

	r.HandleFunc("/api/partners", partnersHandler).Methods("GET")
	r.HandleFunc("/api/partner-owner", partnerOwnerHandler).Methods("GET")
	r.HandleFunc("/api/register-partner", registerPartnerHandler).Methods("POST")
	r.HandleFunc("/api/partner/{id}", partnerDetailsHandler).Methods("GET", "PUT")
	r.HandleFunc("/api/bookings", createBookingHandler).Methods("POST")
	r.HandleFunc("/api/bookings", getBookingsHandler).Methods("GET")
	r.HandleFunc("/api/bookings-client", getBookingsHandlerClientID).Methods("GET")
	r.HandleFunc("/api/bookings/{id}", getBookingsHandlerID).Methods("GET")
	r.HandleFunc("/api/bookings/{id}", updateBookingHandler).Methods("PUT")
	r.HandleFunc("/api/available-times", getAvailableTimeSlotsHandler).Methods("POST")

	r.HandleFunc("/api/announcements/{partner_id}", getAnnouncementsHandler).Methods("GET")
	r.HandleFunc("/api/announcements", createAnnouncementHandler).Methods("POST")
	r.HandleFunc("/api/announcements/{partner_id}/{id}", updateAnnouncementHandler).Methods("PUT")
	r.HandleFunc("/api/announcements/{partner_id}/{id}", deleteAnnouncementHandler).Methods("DELETE")
	r.HandleFunc("/api/announcements/{partner_id}/{id}", getAnnouncementHandler).Methods("GET")
	// Маршруты для управления услугами
	r.HandleFunc("/api/partner_offerings", managePartnerOfferingsHandler).Methods("GET", "POST")
	r.HandleFunc("/api/partner_offerings/{id}", managePartnerOfferingsHandler).Methods("PUT", "DELETE")
	r.HandleFunc("/api/offerings", getOfferingsHandler).Methods("GET")
	// Админ API
	r.HandleFunc("/api/admin/stats", adminStatsHandler).Methods("GET")
	r.HandleFunc("/api/admin/services", adminServicesHandler).Methods("GET")
	r.HandleFunc("/api/admin/approve-service", adminApproveServiceHandler).Methods("POST")
	r.HandleFunc("/api/admin/disapprove-service", adminDisApproveServiceHandler).Methods("POST")
	r.HandleFunc("/api/admin/users", adminUsersHandler).Methods("GET")
	r.HandleFunc("/api/admin/delete-user/{id}", deleteUser).Methods("DELETE")
	//tg bot init
	r.HandleFunc("/api/telegram/init", HandleInitTelegram).Methods("POST")
	// Статические файлы
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
}

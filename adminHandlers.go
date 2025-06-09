package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}
	var stats struct {
		PendingServices int `json:"pending_services"`
		TotalBookings   int `json:"total_bookings"`
		TotalUsers      int `json:"total_users"`
	}

	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE approved = FALSE").Scan(&stats.PendingServices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM bookings").Scan(&stats.TotalBookings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func adminServicesHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}
	rows, err := db.Query(`
		SELECT id, name, address, phone, approved, owner_id 
		FROM services 
		ORDER BY name ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []map[string]interface{}
	for rows.Next() {
		var s Service
		err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.Phone, &s.Approved, &s.OwnerID)
		if err != nil {
			continue
		}
		services = append(services, map[string]interface{}{
			"id":       s.ID,
			"name":     s.Name,
			"address":  s.Address,
			"phone":    s.Phone,
			"approved": s.Approved,
			"owner_id": s.OwnerID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func adminApproveServiceHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	var req struct {
		ServiceID int `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE services SET approved = TRUE WHERE id = $1", req.ServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	adminID, ok := session.Values["user_id"].(int)
	if ok {
		_, err = db.Exec("INSERT INTO admin_logs (admin_id, action) VALUES ($1, $2)",
			adminID, fmt.Sprintf("Approved service ID %d", req.ServiceID))
		if err != nil {
			log.Printf("Error logging admin action: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func adminDisApproveServiceHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	var req struct {
		ServiceID int `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE services SET approved = FALSE WHERE id = $1", req.ServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	adminID, _ := session.Values["user_id"].(int)
	if ok {
		_, err = db.Exec("INSERT INTO admin_logs (admin_id, action) VALUES ($1, $2)",
			adminID, fmt.Sprintf("Disapproved service ID %d", req.ServiceID))
		if err != nil {
			log.Printf("Error logging admin action: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	rows, err := db.Query("SELECT id, name, email, phone, role FROM users ORDER BY id ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role)
		if err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":    u.ID,
			"name":  u.Name,
			"phone": u.Phone,
			"email": u.Email,
			"role":  u.Role,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}
	ID := mux.Vars(r)["id"]

	// Удаляем объявление
	_, err = db.Exec(`
        DELETE FROM users WHERE id = $1`,
		ID)

	if err != nil {
		http.Error(w, "Ошибка при удалении пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Отправляем успешный ответ без содержимого
}

package main

import (
	"html/template"
	"net/http"
)

// Обработчики страниц
// Обработчик index.html
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик partner-register.html
func pagePartnerRegisterHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/partner-register.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик admin.html
func pageAdminHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var user User
	err := db.QueryRow(`
		SELECT role
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.Role)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.Role != "admin" {
		http.Error(w, "У вас нет доступа к этой панели", http.StatusForbidden)
	}

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func pagePartnerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/partner.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id, err := strconv.Atoi(vars["id"])
	// if err != nil {
	// 	http.Error(w, "Неверный ID партнера", http.StatusBadRequest)
	// 	return
	// }
	// session, err := store.Get(r, "session")
	// if err != nil {
	// 	http.Error(w, "Ошибка получения сессии", http.StatusInternalServerError)
	// 	return
	// }
	// userID, ok := session.Values["user_id"].(int)
	// if !ok {
	// 	http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
	// 	return
	// }
	// var partnerOwnerId int
	// err = db.QueryRow(`
	// 	SELECT s.owner_id
	// 	FROM services s
	// 	WHERE s.id = $1 AND s.owner_id = $2
	// `, id, userID).Scan(&partnerOwnerId)
	// if (err == sql.ErrNoRows && ) {
	// 	http.Error(w, "У вас нет доступа к этой панели", http.StatusForbidden)
	// 	return
	// } else if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func clientHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/client.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

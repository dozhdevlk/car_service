package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Ошибка хэширования пароля: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя
	_, err = db.Exec(`
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
	`, req.Name, req.Email, req.Phone, string(hashedPassword), req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Email или телефон уже зарегистрированы", http.StatusBadRequest)
		} else {
			log.Printf("Ошибка сохранения пользователя: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Зарегистрирован пользователь: email=%s, role=%s", req.Email, req.Role)
	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow(`
		SELECT id, name, email, password_hash, role
		FROM users
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err == sql.ErrNoRows {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Ошибка запроса к базе данных: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Создание сессии
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = user.ID
	session.Values["role"] = user.Role
	if err := session.Save(r, w); err != nil {
		log.Printf("Ошибка сохранения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь вошел: email=%s", user.Email)
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("Ошибка удаления сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь вышел и сессия удалена")
	w.WriteHeader(http.StatusOK)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`
		SELECT id, name, email, phone, role, telegram_chat_id
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Tg)

	if err == sql.ErrNoRows {
		log.Printf("Пользователь с ID %d не найден", userID)
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Ошибка получения данных пользователя %d: %v", userID, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Данные пользователя получены: id=%d", userID)

	json.NewEncoder(w).Encode(user)
}

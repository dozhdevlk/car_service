package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func HandleInitTelegram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

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

	token := uuid.New().String()

	result, err := db.Exec(`UPDATE users SET telegram_token = $1 WHERE id = $2`, token, userID)
	if err != nil {
		log.Printf("Ошибка обновления токена для пользователя %d: %v", userID, err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения RowsAffected: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("Пользователь с ID %d не найден", userID)
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	resp := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Ошибка кодирования ответа: %v", err)
	}
}

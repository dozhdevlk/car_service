package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func getAnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	partnerID := mux.Vars(r)["partner_id"]

	// Получение объявлений из базы данных
	rows, err := db.Query(`SELECT id, title, text, image_url FROM announcements WHERE partner_id = $1`, partnerID)
	if err != nil {
		http.Error(w, "Ошибка при получении объявлений", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var announcements []Announcement
	for rows.Next() {
		var ann Announcement
		if err := rows.Scan(&ann.ID, &ann.Title, &ann.Text, &ann.ImageURL); err != nil {
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}
		announcements = append(announcements, ann)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

func createAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	var ann Announcement
	var logoPath string
	file, handler, err := r.FormFile("image_url")
	if err == nil {
		defer file.Close()
		// Создаем папку uploads если нет
		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			os.Mkdir("uploads", 0755)
		}

		// Генерируем уникальное имя файла
		logoPath = "uploads/" + uuid.New().String() + filepath.Ext(handler.Filename)
		f, err := os.Create(logoPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	ann.ImageURL = string(logoPath)
	ann.Text = r.FormValue("text")
	ann.Title = r.FormValue("title")

	// Вставляем данные в базу
	var newID int
	err = db.QueryRow(`
        INSERT INTO announcements (partner_id, title, text, image_url) 
        VALUES ($1, $2, $3, $4) RETURNING id`,
		r.FormValue("partner_id"), r.FormValue("title"), r.FormValue("text"), string(logoPath)).Scan(&newID)

	if err != nil {
		http.Error(w, "Ошибка при добавлении объявления", http.StatusInternalServerError)
		return
	}

	ann.ID = newID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}

func updateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	var ann Announcement
	partnerID := mux.Vars(r)["partner_id"]
	ID := mux.Vars(r)["id"]

	var logoPath string
	file, handler, err := r.FormFile("image_url")
	if err == nil {
		defer file.Close()
		// Создаем папку uploads если нет
		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			os.Mkdir("uploads", 0755)
		}

		// Генерируем уникальное имя файла
		logoPath = "uploads/" + uuid.New().String() + filepath.Ext(handler.Filename)
		f, err := os.Create(logoPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Обновление объявления
	_, err = db.Exec(`
        UPDATE announcements 
        SET title = $1, text = $2, image_url = $3
        WHERE id = $4 AND partner_id = $5`,
		r.FormValue("title"), r.FormValue("text"), logoPath, ID, partnerID)

	ann.ImageURL = string(logoPath)
	ann.Text = r.FormValue("text")
	ann.Title = r.FormValue("title")
	ann.ImageURL = logoPath

	if err != nil {
		http.Error(w, "Ошибка при обновлении объявления", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}

func deleteAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	partnerID := mux.Vars(r)["partner_id"]
	annID := mux.Vars(r)["id"]

	// Удаляем объявление
	_, err := db.Exec(`
        DELETE FROM announcements WHERE id = $1 AND partner_id = $2`,
		annID, partnerID)

	if err != nil {
		http.Error(w, "Ошибка при удалении объявления", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Отправляем успешный ответ без содержимого
}

func getAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	partnerID := mux.Vars(r)["partner_id"]
	id := mux.Vars(r)["id"]

	rows, err := db.Query(`SELECT id, title, text, image_url FROM announcements WHERE partner_id = $1 AND id = $2`, partnerID, id)
	rows.Next()
	if err != nil {
		http.Error(w, "Ошибка при получении объявления", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var ann Announcement
	if err := rows.Scan(&ann.ID, &ann.Title, &ann.Text, &ann.ImageURL); err != nil {
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func managePartnerOfferingsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Получение списка услуг
		partnerIDStr := r.URL.Query().Get("partner_id")
		if partnerIDStr == "" {
			http.Error(w, "Не указан ID сервиса", http.StatusBadRequest)
			return
		}

		partnerID, err := strconv.Atoi(partnerIDStr)
		if err != nil || partnerID <= 0 {
			http.Error(w, "Неверный ID сервиса", http.StatusBadRequest)
			return
		}

		// Изменяем запрос для таблицы partner_offerings
		rows, err := db.Query(`
            SELECT po.id, po.partner_id, o.name, po.price
            FROM partner_offerings po
            JOIN offerings o ON po.offering_id = o.id
            WHERE po.partner_id = $1
        `, partnerID)
		if err != nil {
			http.Error(w, "Не удалось загрузить услуги", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var services []struct {
			ID        int     `json:"id"`
			PartnerID int     `json:"partner_id"`
			Name      string  `json:"name"`
			Price     float64 `json:"price"`
		}

		for rows.Next() {
			var service struct {
				ID        int     `json:"id"`
				PartnerID int     `json:"partner_id"`
				Name      string  `json:"name"`
				Price     float64 `json:"price"`
			}
			if err := rows.Scan(&service.ID, &service.PartnerID, &service.Name, &service.Price); err != nil {
				http.Error(w, "Ошибка чтения услуги", http.StatusInternalServerError)
				return
			}
			services = append(services, service)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(services)
		return
	}

	if r.Method == http.MethodPost {
		// Базовая проверка авторизации
		session, _ := store.Get(r, "session")
		_, ok := session.Values["user_id"].(int)
		if !ok {
			http.Error(w, "Не авторизован", http.StatusUnauthorized)
			return
		}
		var req struct {
			Name      string  `json:"name"`
			Price     float64 `json:"price"`
			PartnerID int     `json:"partner_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Price <= 0 || req.PartnerID <= 0 {
			http.Error(w, "Все поля обязательны", http.StatusBadRequest)
			return
		}

		// Проверяем, существует ли сервис
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)", req.PartnerID).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "Сервис не найден", http.StatusNotFound)
			return
		}

		// Проверяем, существует ли уже услуга с таким названием для данного партнёра
		var existingCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM partner_offerings po
			JOIN offerings o ON po.offering_id = o.id
			WHERE po.partner_id = $1 AND o.name = $2
		`, req.PartnerID, req.Name).Scan(&existingCount)
		if err != nil {
			http.Error(w, "Ошибка проверки существующей услуги", http.StatusInternalServerError)
			return
		}
		if existingCount > 0 {
			http.Error(w, "Услуга с таким названием уже добавлена этим автосервисом", http.StatusBadRequest)
			return
		}

		// Находим или создаём offering
		var offeringID int
		err = db.QueryRow("SELECT id FROM offerings WHERE name = $1", req.Name).Scan(&offeringID)
		if err == sql.ErrNoRows {
			err = db.QueryRow("INSERT INTO offerings (name, description) VALUES ($1, $2) RETURNING id", req.Name, "").Scan(&offeringID)
			if err != nil {
				http.Error(w, "Не удалось создать услугу", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, "Ошибка проверки услуги", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO partner_offerings (partner_id, offering_id, price)
			VALUES ($1, $2, $3)
		`, req.PartnerID, offeringID, req.Price)
		if err != nil {
			http.Error(w, "Не удалось связать услугу с партнером", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         offeringID,
			"partner_id": req.PartnerID,
			"name":       req.Name,
			"price":      req.Price,
		})
		return
	}

	// Извлечение serviceID для PUT и DELETE с использованием mux.Vars
	// Базовая проверка авторизации
	session, _ := store.Get(r, "session")
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	serviceIDStr := vars["id"]
	if serviceIDStr == "" {
		http.Error(w, "Не указан ID услуги", http.StatusBadRequest)
		return
	}
	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil || serviceID <= 0 {
		http.Error(w, "Неверный ID услуги", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPut {
		vars := mux.Vars(r)
		serviceIDStr := vars["id"]
		if serviceIDStr == "" {
			http.Error(w, "Не указан ID услуги", http.StatusBadRequest)
			return
		}
		serviceID, err := strconv.Atoi(serviceIDStr)
		if err != nil || serviceID <= 0 {
			http.Error(w, "Неверный ID услуги", http.StatusBadRequest)
			return
		}

		var request struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		if request.Name == "" || request.Price <= 0 {
			http.Error(w, "Все поля обязательны", http.StatusBadRequest)
			return
		}

		// Обновляем только price в partner_offerings, так как name относится к offerings
		_, err = db.Exec(`
			UPDATE partner_offerings SET price = $1 WHERE id = $2
		`, request.Price, serviceID)
		if err != nil {
			http.Error(w, "Не удалось обновить услугу", http.StatusInternalServerError)
			return
		}

		// Если нужно обновить имя в offerings, добавляем дополнительный запрос
		if request.Name != "" {
			var offeringID int
			err = db.QueryRow(`
				SELECT offering_id FROM partner_offerings WHERE id = $1
			`, serviceID).Scan(&offeringID)
			if err == nil {
				_, err = db.Exec(`
					UPDATE offerings SET name = $1 WHERE id = $2
				`, request.Name, offeringID)
				if err != nil {
					http.Error(w, "Не удалось обновить название услуги", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно обновлена"})
		return
	}

	if r.Method == http.MethodDelete {
		// Удаление услуги
		result, err := db.Exec("DELETE FROM partner_offerings WHERE id = $1", serviceID)
		if err != nil {
			http.Error(w, "Не удалось удалить услугу", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Ошибка при проверке удаления", http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			http.Error(w, "Услуга не найдена", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно удалена"})
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
func getOfferingsHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем все уникальные услуги
	rows, err := db.Query(`
        SELECT DISTINCT o.id, o.name
        FROM offerings o
        JOIN partner_offerings po ON o.id = po.offering_id
        ORDER BY o.name
    `)
	if err != nil {
		http.Error(w, "Не удалось загрузить услуги", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var offerings []OfferingResponse
	for rows.Next() {
		var offering OfferingResponse
		if err := rows.Scan(&offering.ID, &offering.Name); err != nil {
			http.Error(w, "Ошибка чтения услуги", http.StatusInternalServerError)
			return
		}

		// Получаем всех партнёров для данной услуги
		partnerRows, err := db.Query(`
            SELECT s.id, s.name, po.price, COALESCE(po.image_url, '')
            FROM partner_offerings po
            JOIN services s ON po.partner_id = s.id
            WHERE po.offering_id = $1
        `, offering.ID)
		if err != nil {
			http.Error(w, "Не удалось загрузить партнёров для услуги", http.StatusInternalServerError)
			return
		}
		defer partnerRows.Close()

		var partners []PartnerOfferingInfo
		for partnerRows.Next() {
			var partner PartnerOfferingInfo
			if err := partnerRows.Scan(&partner.ID, &partner.Name, &partner.Price, &partner.ImageURL); err != nil {
				http.Error(w, "Ошибка чтения данных партнёра", http.StatusInternalServerError)
				return
			}
			partners = append(partners, partner)
		}

		offering.Partners = partners
		offerings = append(offerings, offering)
	}

	// Проверяем, есть ли услуги
	if len(offerings) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]OfferingResponse{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offerings)
}

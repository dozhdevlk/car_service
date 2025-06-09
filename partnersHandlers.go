package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func partnersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name
		FROM services s
		JOIN users u ON s.owner_id = u.id
		WHERE s.approved = TRUE
		ORDER BY s.name
	`)
	if err != nil {
		log.Printf("Ошибка получения списка партнеров: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var partners []map[string]interface{}
	for rows.Next() {
		var id int
		var name, address, phone, owner string
		var logoPath sql.NullString
		var latitude, longitude float64

		err := rows.Scan(&id, &name, &address, &phone, &logoPath, &latitude, &longitude, &owner)
		if err != nil {
			continue
		}

		partner := map[string]interface{}{
			"id":        id,
			"name":      name,
			"address":   address,
			"phone":     phone,
			"logoPath":  logoPath.String,
			"longitude": longitude,
			"latitude":  latitude,
			"owner":     owner,
		}
		partners = append(partners, partner)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(partners)
}
func partnerOwnerHandler(w http.ResponseWriter, r *http.Request) {
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
	var id int
	err = db.QueryRow(`
            SELECT s.id
            FROM services s
            WHERE s.owner_id = $1
        `, userID).Scan(&id)
	if err != nil {
		http.Error(w, "Сервис не найден", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(id)
}

func registerPartnerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	// Проверка email
	email := r.FormValue("ownerEmail")
	var emailExists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&emailExists)

	if err != nil {
		log.Printf("Ошибка проверки email: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if emailExists {
		http.Error(w, "Email уже используется", http.StatusBadRequest)
		return
	}

	ownerName := r.FormValue("ownerName")
	ownerPhone := r.FormValue("ownerPhone")
	ownerPassword := r.FormValue("ownerPassword")
	serviceName := r.FormValue("serviceName")
	serviceAddress := r.FormValue("serviceAddress")
	servicePhone := r.FormValue("servicePhone")
	if ownerName == "" || ownerPhone == "" || ownerPassword == "" || serviceName == "" || serviceAddress == "" || servicePhone == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	var logoPath string
	file, handler, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()

		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			if err := os.Mkdir("uploads", 0755); err != nil {
				log.Printf("Ошибка создания директории uploads: %v", err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}

		logoPath = "uploads/" + uuid.New().String() + filepath.Ext(handler.Filename)
		f, err := os.Create(logoPath)
		if err != nil {
			log.Printf("Ошибка создания файла логотипа: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			log.Printf("Ошибка копирования файла логотипа: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(ownerPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Ошибка хэширования пароля: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	var ownerID int
	err = db.QueryRow(`
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, 'admin_service')
		RETURNING id
	`, r.FormValue("ownerName"), email, r.FormValue("ownerPhone"), string(hashedPassword)).Scan(&ownerID)
	if err != nil {
		log.Printf("Ошибка регистрации владельца: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	lat, lon, err := geocodeAddress(r.FormValue("serviceAddress"))
	if err != nil {
		log.Printf("Геокодирование не удалось: %v", err)
	}
	workHours := map[string]map[string]string{
		"mon": {"from": r.FormValue("hours_mon_from"), "to": r.FormValue("hours_mon_to")},
		"tue": {"from": r.FormValue("hours_tue_from"), "to": r.FormValue("hours_tue_to")},
		"wed": {"from": r.FormValue("hours_wed_from"), "to": r.FormValue("hours_wed_to")},
		"thu": {"from": r.FormValue("hours_thu_from"), "to": r.FormValue("hours_thu_to")},
		"fri": {"from": r.FormValue("hours_fri_from"), "to": r.FormValue("hours_fri_to")},
		"sat": {"from": r.FormValue("hours_sat_from"), "to": r.FormValue("hours_sat_to")},
		"sun": {"from": r.FormValue("hours_sun_from"), "to": r.FormValue("hours_sun_to")},
	}
	workHoursJSON, err := json.Marshal(workHours)
	if err != nil {
		log.Printf("Ошибка сериализации рабочих часов: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec(`
	INSERT INTO services (name, address, phone, logo_path, owner_id, latitude, longitude, working_hours)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`,
		serviceName,
		serviceAddress,
		servicePhone,
		logoPath,
		ownerID,
		lat,
		lon,
		workHoursJSON,
	)
	if err != nil {
		log.Printf("Ошибка регистрации сервиса: %v", err)
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
		log.Printf("Сервис не был создан для owner_id %d", ownerID)
		http.Error(w, "Не удалось создать сервис", http.StatusInternalServerError)
		return
	}

	log.Printf("Сервис успешно зарегистрирован: name=%s, owner_id=%d", serviceName, ownerID)
	w.WriteHeader(http.StatusCreated)
}

func partnerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	var partner Partner
	switch r.Method {
	case http.MethodGet:
		var mapBlock sql.NullString
		var reviewsBlock sql.NullString

		var workingHoursJson []byte
		err = db.QueryRow(`
			SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name, s.owner_id, s.description, s.working_hours, s.map, s.reviews
			FROM services s
			JOIN users u ON s.owner_id = u.id
			WHERE s.id = $1 AND s.approved = TRUE
		`, id).Scan(&partner.ID, &partner.Name, &partner.Address, &partner.Phone, &partner.LogoPath, &partner.Latitude, &partner.Longitude, &partner.Owner, &partner.Owner_id, &partner.Description, &workingHoursJson, &mapBlock, &reviewsBlock)
		if err == sql.ErrNoRows {
			http.Error(w, "Партнер не найден", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Ошибка получения данных партнера %d: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		if mapBlock.Valid {
			partner.MapBlock = mapBlock.String
		} else {
			partner.MapBlock = ""
		}
		if reviewsBlock.Valid {
			partner.ReviewsBlock = reviewsBlock.String
		} else {
			partner.ReviewsBlock = ""
		}

		if len(workingHoursJson) > 0 {
			err = json.Unmarshal(workingHoursJson, &partner.WorkingHours)
			if err != nil {
				log.Printf("Ошибка парсинга рабочих часов для партнера %d: %v", id, err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(partner); err != nil {
			log.Printf("Ошибка кодирования ответа: %v", err)
		}
	case http.MethodPut:
		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("Ошибка получения сессии: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		userID, ok := session.Values["user_id"].(int)
		if !ok {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		err = db.QueryRow(`
			SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name, s.owner_id, s.description, s.working_hours
			FROM services s
			JOIN users u ON s.owner_id = u.id
			WHERE s.id = $1 AND s.approved = TRUE
		`, id).Scan(&partner.ID, &partner.Name, &partner.Address, &partner.Phone, &partner.LogoPath, &partner.Latitude, &partner.Longitude, &partner.Owner, &partner.Owner_id, &partner.Description, &partner.WorkingHours)
		if err == sql.ErrNoRows {
			http.Error(w, "Партнер не найден", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Ошибка получения данных партнера %d: %v", id, err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		if partner.Owner_id != userID {
			http.Error(w, "У вас нет прав для редактирования этого сервиса", http.StatusForbidden)
			return
		}

		var updateData struct {
			Name         string               `json:"name"`
			Phone        string               `json:"phone"`
			Description  string               `json:"description"`
			WorkingHours map[string][2]string `json:"working_hours"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
			return
		}

		// Проверяем обязательные поля
		if updateData.Name == "" || updateData.Phone == "" {
			http.Error(w, "Имя и телефон обязательны", http.StatusBadRequest)
			return
		}

		workHoursJSON, err := json.Marshal(updateData.WorkingHours)
		if err != nil {
			log.Printf("Ошибка сериализации рабочих часов: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		result, err := db.Exec(`
			UPDATE services
			SET name = $1, phone = $2, description = $3, working_hours = $4
			WHERE id = $5
		`, updateData.Name, updateData.Phone, updateData.Description, workHoursJSON, id)
		if err != nil {
			log.Printf("Ошибка обновления партнера %d: %v", id, err)
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
			http.Error(w, "Партнер не найден", http.StatusNotFound)
			return
		}

		log.Printf("Партнер успешно обновлен: id=%d", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Партнер успешно обновлен",
		})

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

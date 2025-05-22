package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	db    *sql.DB
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)
var apiKey string

type User struct {
	ID           int
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
}

type Service struct {
	ID       int
	Name     string
	Address  string
	Phone    string
	Approved bool
	OwnerID  int
}

type Booking struct {
	ID             int    `json:"id"`
	PartnerID      int    `json:"partner_id"`
	PartnerName    string `json:"partner_name"`
	PartnerPhone   string `json:"partner_phone"`
	PartnerAddress string `json:"partner_address"`
	UserID         int    `json:"user_id"`
	UserName       string `json:"user_name"`
	UserPhone      string `json:"user_phone"`
	UserEmail      string `json:"user_email"`
	BookingDate    string `json:"booking_date"`
	BookingTime    string `json:"booking_time"`
	Status         string `json:"status"`
}

type Offering struct {
	ID          int
	Name        string
	Description string
}

type PartnerOffering struct {
	ID         int
	PartnerID  int
	OfferingID int
	Price      float64
	ImageURL   string
}

// Структура для ответа API
type OfferingResponse struct {
	ID       int                   `json:"id"`
	Name     string                `json:"name"`
	Partners []PartnerOfferingInfo `json:"partners"`
}

type PartnerOfferingInfo struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url"`
}
type Announcement struct {
	ID        int    `json:"id"`
	PartnerID int    `json:"partner_id"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	ImageURL  string `json:"image_url"`
}

func main() {

	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,                    // Безопасность: cookie будет доступен только через HTTP
		Secure:   false,                   // Включить, если используется HTTPS
		SameSite: http.SameSiteStrictMode, // Защита от CSRF атак
	}

	apiKey = os.Getenv("API_KEY_GEOCODE")
	// Инициализация БД
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблиц при первом запуске
	initDB()

	// Роутер
	r := mux.NewRouter()
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
	// Статические файлы
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}

func geocodeAddress(address string) (float64, float64, error) {
	url := fmt.Sprintf(
		"https://geocode-maps.yandex.ru/1.x/?format=json&apikey=%s&geocode=%s",
		apiKey,
		url.QueryEscape(address),
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка запроса к геокодеру: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Парсим JSON ответ
	var result struct {
		Response struct {
			GeoObjectCollection struct {
				FeatureMember []struct {
					GeoObject struct {
						Point struct {
							Pos string `json:"pos"`
						} `json:"Point"`
					} `json:"GeoObject"`
				} `json:"featureMember"`
			} `json:"GeoObjectCollection"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	// Проверяем наличие результатов
	if len(result.Response.GeoObjectCollection.FeatureMember) == 0 {
		return 0, 0, fmt.Errorf("адрес не найден")
	}

	// Разбираем координаты (долгота и широта)
	coords := strings.Split(result.Response.GeoObjectCollection.FeatureMember[0].GeoObject.Point.Pos, " ")
	if len(coords) != 2 {
		return 0, 0, fmt.Errorf("неверный формат координат")
	}

	lon, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка парсинга долготы: %v", err)
	}

	lat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка парсинга широты: %v", err)
	}

	return lat, lon, nil
}

// Создание таблиц если не существуют
func initDB() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)
`)
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS services (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            address TEXT NOT NULL,
            phone TEXT NOT NULL,
            logo_path TEXT,
            owner_id INTEGER NOT NULL,
            FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
			latitude DECIMAL(10, 7),
			longitude DECIMAL(10, 7),
			approved BOOLEAN DEFAULT FALSE,
			description TEXT DEFAULT 'Отсутствует',
			working_hours JSONB,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		log.Fatal("Error creating services table:", err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_services_owner_id ON services(owner_id)")
	if err != nil {
		log.Fatal("Error creating index:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS admin_logs (
			id SERIAL PRIMARY KEY,
			admin_id INT REFERENCES users(id),
			action TEXT,
			created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		log.Fatal("Error creating admin_logs table:", err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		partner_id INTEGER,
		user_id INTEGER,
		booking_date TEXT,
		booking_time TEXT,
		status TEXT,
		FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE
	);`)
	if err != nil {
		log.Fatal("Error creating bookings table:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS offerings (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		)
	`)
	if err != nil {
		log.Fatal("Error creating offerings table:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS partner_offerings (
			id SERIAL PRIMARY KEY,
			partner_id INTEGER NOT NULL,
			offering_id INTEGER NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			image_url TEXT,
			FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE,
			FOREIGN KEY (offering_id) REFERENCES offerings(id)
		)
	`)
	if err != nil {
		log.Fatal("Error creating partner_offerings table:", err)
	}
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS announcements (
        id SERIAL PRIMARY KEY,
        partner_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        text TEXT NOT NULL,
        image_url TEXT,
        FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE
    )
`)
	if err != nil {
		log.Fatal("Error creating announcements table:", err)
	}

}

// Обработчики API
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя
	_, err = db.Exec(`
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
	`, req.Name, req.Email, req.Phone, string(hashedPassword), req.Role)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow(`
		SELECT id, name, email, password_hash, role
		FROM users
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Создание сессии
	session, _ := store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	json.NewEncoder(w).Encode(user)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err := db.QueryRow(`
		SELECT id, name, email, phone, role
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func partnersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name
		FROM services s
		JOIN users u ON s.owner_id = u.id
		WHERE s.approved = TRUE
		ORDER BY s.name
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func registerPartnerHandler(w http.ResponseWriter, r *http.Request) {
	// Ограничение 10MB для загрузки
	r.ParseMultipartForm(10 << 20)

	// Проверка email
	email := r.FormValue("ownerEmail")
	var emailExists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, email).Scan(&emailExists)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if emailExists {
		http.Error(w, "Email уже используется", http.StatusBadRequest)
		return
	}

	// Обработка логотипа
	var logoPath string
	file, handler, err := r.FormFile("logo")
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

	// Регистрация владельца
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(r.FormValue("ownerPassword")),
		bcrypt.DefaultCost,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ownerID int
	err = db.QueryRow(`
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, 'admin_service')
		RETURNING id
	`, r.FormValue("ownerName"), email, r.FormValue("ownerPhone"), string(hashedPassword)).Scan(&ownerID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//получение координат
	lat, lon, err := geocodeAddress(r.FormValue("serviceAddress"))
	if err != nil {
		log.Printf("Геокодирование не удалось: %v", err)
		// Можно продолжить без координат или вернуть ошибку
	}
	// Создаем структуру для хранения рабочего времени
	workHours := map[string]map[string]string{
		"mon": {"from": r.FormValue("hours_mon_from"), "to": r.FormValue("hours_mon_to")},
		"tue": {"from": r.FormValue("hours_tue_from"), "to": r.FormValue("hours_tue_to")},
		"wed": {"from": r.FormValue("hours_wed_from"), "to": r.FormValue("hours_wed_to")},
		"thu": {"from": r.FormValue("hours_thu_from"), "to": r.FormValue("hours_thu_to")},
		"fri": {"from": r.FormValue("hours_fri_from"), "to": r.FormValue("hours_fri_to")},
		"sat": {"from": r.FormValue("hours_sat_from"), "to": r.FormValue("hours_sat_to")},
		"sun": {"from": r.FormValue("hours_sun_from"), "to": r.FormValue("hours_sun_to")},
	}
	// Сериализуем рабочие часы в JSON
	workHoursJSON, err := json.Marshal(workHours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Регистрация сервиса
	_, err = db.Exec(`
	INSERT INTO services (name, address, phone, logo_path, owner_id, latitude, longitude, working_hours)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`,
		r.FormValue("serviceName"),
		r.FormValue("serviceAddress"),
		r.FormValue("servicePhone"),
		string(logoPath),
		ownerID,
		lat,
		lon,
		workHoursJSON,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func partnerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	var partner struct {
		ID           int
		Name         string
		Address      string
		Phone        string
		LogoPath     sql.NullString
		Latitude     float64
		Longitude    float64
		Owner        string
		Owner_id     int
		Description  string
		WorkingHours map[string]map[string]string // Структура для рабочих часов
	}
	// Обработка в зависимости от метода
	switch r.Method {
	case http.MethodGet:
		var workingHoursJson []byte
		err = db.QueryRow(`
			SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name, s.owner_id, s.description, s.working_hours
			FROM services s
			JOIN users u ON s.owner_id = u.id
			WHERE s.id = $1 AND s.approved = TRUE
		`, id).Scan(&partner.ID, &partner.Name, &partner.Address, &partner.Phone, &partner.LogoPath, &partner.Latitude, &partner.Longitude, &partner.Owner, &partner.Owner_id, &partner.Description, &workingHoursJson)

		// Преобразование данных о working_hours
		if err == nil && partner.WorkingHours == nil {
			err := db.QueryRow(`SELECT working_hours FROM services WHERE id = $1`, id).Scan(&workingHoursJson)
			if err != nil {
				http.Error(w, "Error fetching working hours", http.StatusInternalServerError)
				return
			}

			// Парсим JSON в рабочие часы
			if len(workingHoursJson) > 0 {
				err = json.Unmarshal(workingHoursJson, &partner.WorkingHours)
				if err != nil {
					http.Error(w, "Error parsing working hours", http.StatusInternalServerError)
					return
				}
			}
		}

		if err == sql.ErrNoRows {
			http.Error(w, "Partner not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            partner.ID,
			"name":          partner.Name,
			"address":       partner.Address,
			"phone":         partner.Phone,
			"logoPath":      partner.LogoPath.String,
			"latitude":      partner.Latitude,
			"longitude":     partner.Longitude,
			"owner":         partner.Owner,
			"owner_id":      partner.Owner_id,
			"description":   partner.Description,
			"working_hours": partner.WorkingHours, // Добавляем рабочие часы в ответ
		})
	case http.MethodPut:
		// Получаем userID из сессии
		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, "Ошибка получения сессии", http.StatusInternalServerError)
			return
		}

		userID, ok := session.Values["user_id"].(int)
		if !ok {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		// Загружаем текущие данные сервиса
		err = db.QueryRow(`
			SELECT s.id, s.name, s.address, s.phone, s.logo_path, s.latitude, s.longitude, u.name as owner_name, s.owner_id, s.description, s.working_hours
			FROM services s
			JOIN users u ON s.owner_id = u.id
			WHERE s.id = $1 AND s.approved = TRUE
		`, id).Scan(&partner.ID, &partner.Name, &partner.Address, &partner.Phone, &partner.LogoPath, &partner.Latitude, &partner.Longitude, &partner.Owner, &partner.Owner_id, &partner.Description, &partner.WorkingHours)
		if err == sql.ErrNoRows {
			http.Error(w, "Partner not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Проверка прав
		if partner.Owner_id != userID {
			http.Error(w, "У вас нет прав для редактирования этого сервиса", http.StatusForbidden)
			return
		}

		// Парсим тело запроса
		var updateData struct {
			Name         string               `json:"name"`
			Phone        string               `json:"phone"`
			Description  string               `json:"description"`
			WorkingHours map[string][2]string `json:"working_hours"` // Поле рабочих часов
		}
		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Проверяем обязательные поля
		if updateData.Name == "" || updateData.Phone == "" {
			http.Error(w, "Name and phone are required", http.StatusBadRequest)
			return
		}

		// Обновляем данные в базе данных
		_, err = db.Exec(`
			UPDATE services
			SET name = $1, phone = $2, description = $3, working_hours = $4
			WHERE id = $5
		`, updateData.Name, updateData.Phone, updateData.Description, updateData.WorkingHours, id)
		if err != nil {
			http.Error(w, "Failed to update partner info", http.StatusInternalServerError)
			return
		}

		// Возвращаем успешный ответ
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Partner information updated successfully",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Admin API handlers
func adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		PendingServices int `json:"pending_services"`
		TotalBookings   int `json:"total_bookings"`
		TotalUsers      int `json:"total_users"`
	}

	err := db.QueryRow("SELECT COUNT(*) FROM services WHERE approved = FALSE").Scan(&stats.PendingServices)
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
	var req struct {
		ServiceID int `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE services SET approved = TRUE WHERE id = $1", req.ServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session")
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
	var req struct {
		ServiceID int `json:"service_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE services SET approved = FALSE WHERE id = $1", req.ServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session")
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
func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
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
	ID := mux.Vars(r)["id"]

	// Удаляем объявление
	_, err := db.Exec(`
        DELETE FROM users WHERE id = $1`,
		ID)

	fmt.Println(err)
	if err != nil {
		http.Error(w, "Ошибка при удалении пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Отправляем успешный ответ без содержимого
}

// API добавленеия услуг
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

func createPartnerOfferingHandler(w http.ResponseWriter, r *http.Request) {
	// Ограничение 10MB для загрузки
	r.ParseMultipartForm(10 << 20)

	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверка, является ли пользователь владельцем сервиса
	var partnerID int
	err := db.QueryRow("SELECT id FROM services WHERE owner_id = $1 AND approved = TRUE LIMIT 1", userID).Scan(&partnerID)
	if err != nil {
		http.Error(w, "Вы не являетесь владельцем одобренного сервиса", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "Неверный формат цены", http.StatusBadRequest)
		return
	}

	// Проверка, существует ли уже услуга с таким названием у данного партнёра
	var existingOffering int
	err = db.QueryRow(`
        SELECT COUNT(*)
        FROM partner_offerings po
        JOIN offerings o ON po.offering_id = o.id
        WHERE po.partner_id = $1 AND o.name = $2
    `, partnerID, name).Scan(&existingOffering)
	if err != nil {
		http.Error(w, "Ошибка проверки существующей услуги", http.StatusInternalServerError)
		return
	}
	if existingOffering > 0 {
		http.Error(w, "Услуга с таким названием уже добавлена этим автосервисом", http.StatusBadRequest)
		return
	}

	var offeringID int
	err = db.QueryRow("SELECT id FROM offerings WHERE name = $1", name).Scan(&offeringID)
	if err == sql.ErrNoRows {
		// Если услуги с таким названием нет, создаём новую
		err = db.QueryRow(
			"INSERT INTO offerings (name, description) VALUES ($1, $2) RETURNING id",
			name, "",
		).Scan(&offeringID)
		if err != nil {
			http.Error(w, "Не удалось создать услугу", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Ошибка проверки услуги", http.StatusInternalServerError)
		return
	}

	var imageURL string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		if _, err := os.Stat("images"); os.IsNotExist(err) {
			os.Mkdir("images", 0755)
		}

		imageURL = "images/" + uuid.New().String() + filepath.Ext(handler.Filename)
		f, err := os.Create(imageURL)
		if err != nil {
			http.Error(w, "Не удалось сохранить изображение", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			http.Error(w, "Ошибка копирования изображения", http.StatusInternalServerError)
			return
		}
	}

	_, err = db.Exec(`
		INSERT INTO partner_offerings (partner_id, offering_id, price, image_url)
		VALUES ($1, $2, $3, $4)
	`, partnerID, offeringID, price, imageURL)
	if err != nil {
		http.Error(w, "Не удалось связать услугу с партнером", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно добавлена"})
}

// Bookings Api handlers
func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var booking Booking

	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if booking.BookingDate == "" || booking.BookingTime == "" {
		http.Error(w, "Дата и время обязательны", http.StatusBadRequest)
		return
	}

	var partnerExists int
	err := db.QueryRow("SELECT COUNT(*) FROM services WHERE id = $1", booking.PartnerID).Scan(&partnerExists)

	if err != nil || partnerExists == 0 {
		http.Error(w, "Партнер не найден", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var existingBooking int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookings 
		WHERE partner_id = $1 AND booking_date = $2 AND booking_time = $3 AND status != 'canceled'`,
		booking.PartnerID, booking.BookingDate, booking.BookingTime).Scan(&existingBooking)

	if err != nil {
		http.Error(w, "Ошибка проверки существующей записи", http.StatusInternalServerError)
		return
	}
	if existingBooking > 0 {
		http.Error(w, "Этот временной слот уже занят", http.StatusConflict)
		return
	}

	result, err := db.Exec(`
		INSERT INTO bookings (partner_id, user_id, booking_date, booking_time, status) 
		VALUES ($1, $2, $3, $4, $5)`,
		booking.PartnerID, userID, booking.BookingDate, booking.BookingTime, booking.Status,
	)
	if err != nil {
		http.Error(w, "Не удалось создать запись", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	booking.ID = int(id)
	booking.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}
func getBookingsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
    SELECT 
        b.id, 
        b.partner_id, 
        b.user_id, 
        b.booking_date, 
        b.booking_time, 
        b.status, 
        u.name AS user_name, 
        u.email AS user_email,
        s.name AS partner_name,
        s.address AS partner_address,
		s.phone AS partner_phone
    FROM bookings b
    JOIN users u ON b.user_id = u.id
    JOIN services s ON b.partner_id = s.id
`)
	if err != nil {
		http.Error(w, "Не удалось загрузить записи", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.ID, &booking.PartnerID, &booking.UserID, &booking.BookingDate, &booking.BookingTime, &booking.Status, &booking.UserName, &booking.UserEmail, &booking.PartnerName, &booking.PartnerAddress, &booking.PartnerPhone); err != nil {
			http.Error(w, "Ошибка чтения записи", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
func getBookingsHandlerClientID(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, "Ошибка получения сессии", http.StatusInternalServerError)
		return
	}
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	rows, err := db.Query(`
    SELECT 
        b.id, 
        b.partner_id, 
        b.user_id, 
        b.booking_date, 
        b.booking_time, 
        b.status, 
        u.name AS user_name, 
        u.email AS user_email,
        s.name AS partner_name,
        s.address AS partner_address,
		s.phone AS partner_phone
    FROM bookings b
    JOIN users u ON b.user_id = u.id
    JOIN services s ON b.partner_id = s.id
	WHERE b.user_id = $1"
`, userID)
	if err != nil {
		http.Error(w, "Не удалось загрузить записи", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.ID, &booking.PartnerID, &booking.UserID, &booking.BookingDate, &booking.BookingTime, &booking.Status, &booking.UserName, &booking.UserEmail, &booking.PartnerName, &booking.PartnerAddress, &booking.PartnerPhone); err != nil {
			http.Error(w, "Ошибка чтения записи", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)

}
func getBookingsHandlerID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
    SELECT 
        b.id, 
        b.partner_id, 
        b.user_id, 
        b.booking_date, 
        b.booking_time, 
        b.status, 
        u.name AS user_name, 
        u.email AS user_email,
        s.name AS partner_name,
        s.address AS partner_address,
		s.phone AS partner_phone
    FROM bookings b
    JOIN users u ON b.user_id = u.id
    JOIN services s ON b.partner_id = s.id
	WHERE b.partner_id = $1"
`, id)
	if err != nil {
		http.Error(w, "Не удалось загрузить записи", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.ID, &booking.PartnerID, &booking.UserID, &booking.BookingDate, &booking.BookingTime, &booking.Status, &booking.UserName, &booking.UserEmail, &booking.PartnerName, &booking.PartnerAddress, &booking.PartnerPhone); err != nil {
			http.Error(w, "Ошибка чтения записи", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
func updateBookingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updateData struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if updateData.Status != "confirmed" && updateData.Status != "canceled" {
		http.Error(w, "Неверный статус", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE bookings SET status = $1 WHERE id = $2", updateData.Status, id)
	if err != nil {
		http.Error(w, "Не удалось обновить запись", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Запись успешно обновлена"})
}

func partnerOwnerHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	var id int
	err := db.QueryRow(`
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

func managePartnerOfferingsHandler(w http.ResponseWriter, r *http.Request) {
	// Базовая проверка авторизации
	session, _ := store.Get(r, "session")
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

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

func getAvailableTimeSlotsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PartnerID   int    `json:"partner_id"`
		BookingDate string `json:"booking_date"`
	}

	// Чтение данных из запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Получаем все занятые временные слоты для указанной даты и автосервиса
	rows, err := db.Query(`
		SELECT booking_time 
		FROM bookings 
		WHERE partner_id = $1 AND booking_date = $2 AND status != 'canceled'`,
		req.PartnerID, req.BookingDate)
	if err != nil {
		http.Error(w, "Не удалось получить занятые временные слоты", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создаем список занятых временных слотов
	var occupiedSlots []string
	for rows.Next() {
		var timeSlot string
		err := rows.Scan(&timeSlot)
		if err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		occupiedSlots = append(occupiedSlots, timeSlot)
	}

	// Создаем список всех временных промежутков для выбранной даты
	availableSlots := getAvailableTimeSlots(req.BookingDate, req.PartnerID, occupiedSlots)

	// Отправляем доступные временные промежутки
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableSlots)
}

// Функция для генерации всех доступных временных промежутков
func getAvailableTimeSlots(bookingDate string, partnerID int, occupiedSlots []string) []string {
	// Получаем рабочие часы автосервиса
	var workingHours string
	err := db.QueryRow(`SELECT working_hours FROM services WHERE id = $1`, partnerID).Scan(&workingHours)
	if err != nil {
		log.Fatal(err)
	}

	// Парсим рабочие часы из JSON
	var hours map[string]map[string]string
	err = json.Unmarshal([]byte(workingHours), &hours)
	if err != nil {
		log.Fatal(err)
	}

	// Формируем список всех временных промежутков
	var timeSlots []string
	for dayKey, times := range hours {
		// Проверяем, совпадает ли день с днем недели на выбранной дате
		selectedDate := parseDate(bookingDate)
		dayOfWeek := selectedDate.Weekday()
		if dayKey != getDayKey(dayOfWeek) {
			continue
		}

		// Генерация всех временных интервалов по дням (каждые 15 минут)
		startTime := times["from"] // "08:00"
		endTime := times["to"]     // "17:00"

		fmt.Println(startTime)
		fmt.Println(endTime)

		if startTime == endTime && startTime == "00:00" {
			return timeSlots
		}

		// Парсим время начала и окончания в часовые и минутные компоненты
		startHour, startMinute, err := parseTime(startTime)
		if err != nil {
			log.Println("Ошибка парсинга времени начала: ", err)
			continue
		}

		endHour, endMinute, err := parseTime(endTime)
		if err != nil {
			log.Println("Ошибка парсинга времени окончания: ", err)
			continue
		}

		// Генерация всех временных интервалов с шагом 15 минут
		currentHour, currentMinute := startHour, startMinute
		for {
			// Формируем строку времени в формате "HH:MM"
			timeSlot := fmt.Sprintf("%02d:%02d", currentHour, currentMinute)

			// Проверяем, заняты ли слоты, и если нет, добавляем в список
			if !isSlotOccupied(timeSlot, occupiedSlots) {
				timeSlots = append(timeSlots, timeSlot)
			}

			// Если мы достигли времени окончания, выходим из цикла
			if currentHour == endHour && currentMinute == endMinute {
				break
			}

			// Увеличиваем время на 15 минут
			currentMinute += 15
			if currentMinute == 60 {
				currentMinute = 0
				currentHour++
			}
		}
	}

	// Возвращаем все доступные временные промежутки
	return timeSlots
}

// Функция для парсинга времени в формате "HH:MM" в часы и минуты
func parseTime(t string) (int, int, error) {
	var hour, minute int
	_, err := fmt.Sscanf(t, "%d:%d", &hour, &minute)
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось распарсить время: %s", t)
	}
	return hour, minute, nil
}

// Функция для проверки, занято ли время
func isSlotOccupied(timeSlot string, occupiedSlots []string) bool {
	for _, slot := range occupiedSlots {
		if timeSlot == slot {
			return true
		}
	}
	return false
}

// Функция для получения дня недели из числа (0 - воскресенье, 1 - понедельник и т.д.)
func getDayKey(dayOfWeek time.Weekday) string {
	days := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	return days[dayOfWeek]
}

// Функция для парсинга строки даты в объект Date
func parseDate(dateStr string) time.Time {
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Fatal("Ошибка парсинга даты: ", err)
	}
	return parsedDate
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
		fmt.Println(err)
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}

// Получение всех объявлений для партнера
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

// Создание нового объявления
func createAnnouncementHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.FormValue("partner_id"), r.FormValue("title"), r.FormValue("text"))
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
		fmt.Println(err)
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

// Обновление объявления
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
		fmt.Println(err)
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

// Удаление объявления
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
	userRole, ok := session.Values["role"].(string)
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}
	if userRole != "client" {
		http.Error(w, "Только клиенты могут создавать бронирования", http.StatusForbidden)
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

	var bookingID int
	err = db.QueryRow(`
		INSERT INTO bookings (partner_id, user_id, booking_date, booking_time, status) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		booking.PartnerID, userID, booking.BookingDate, booking.BookingTime, booking.Status,
	).Scan(&bookingID)
	if err != nil {
		http.Error(w, "Не удалось создать запись", http.StatusInternalServerError)
		return
	}

	// var telegram_chat_id *int64
	// err = db.QueryRow(`
	// SELECT u.telegram_chat_id
	// JOIN services s ON s.owner_id = u.id
	// WHERE s.id = $1
	// `, booking.PartnerID).Scan(&telegram_chat_id)

	var user User
	err = db.QueryRow(`
    SELECT 
        u.name,
		u.phone,
        u.email
    FROM bookings b
    JOIN users u ON b.user_id = u.id
    JOIN services s ON b.partner_id = s.id
	WHERE b.id = $1
	`, bookingID).Scan(&user.Name, &user.Phone, &user.Email)
	if err != nil {
		log.Printf("Ошибка получения созданной записи: %v", err)
	}
	message := EscapeMarkdownV2(fmt.Sprintf(
		"Новая запись №%d:\n\n📅*Дата:* %s\n🕒*Время:* %s\n\n*Клиент(id: %d)*\n\n*Имя:* %s\n*Email:* %s\n*Номер телефона:* %s\n\n*Cтатус: %s*",
		bookingID,
		booking.BookingDate,
		booking.BookingTime,
		booking.UserID,
		user.Name,
		user.Email,
		user.Phone,
		booking.Status,
	))
	SendTelegramNotification(db, booking.UserID, message)

	booking.ID = bookingID
	booking.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
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
				WHERE b.user_id = $1
		`, userID)
	fmt.Println(err)
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
		http.Error(w, "Несуществующий ID партнера", http.StatusBadRequest)
		return
	}
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	userRole, ok := session.Values["role"].(string)
	if !ok || (userRole != "admin" && userRole != "admin_service") {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
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
				WHERE b.partner_id = $1
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
	if (userRole != "admin") && (userRole != "admin_service") {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	var updateData struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
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

	var booking Booking
	err = db.QueryRow(`
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
		WHERE b.id = $1
	`, id).Scan(
		&booking.ID,
		&booking.PartnerID,
		&booking.UserID,
		&booking.BookingDate,
		&booking.BookingTime,
		&booking.Status,
		&booking.UserName,
		&booking.UserEmail,
		&booking.PartnerName,
		&booking.PartnerAddress,
		&booking.PartnerPhone,
	)
	if err != nil {
		log.Println("Ошибка при получении информации о записи:", err)
	} else {
		message := EscapeMarkdownV2(fmt.Sprintf(
			"Обновление записи №%d:\n\n🚗*СТО:* %s\n📍*Адрес:* %s\n📅*Дата:* %s\n🕒*Время:* %s\n\n*Новый статус: %s*",
			booking.ID,
			booking.PartnerName,
			booking.PartnerAddress,
			booking.BookingDate,
			booking.BookingTime,
			booking.Status,
		))

		SendTelegramNotification(db, booking.UserID, message)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Запись успешно обновлена"})
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

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

func EscapeMarkdownV2(text string) string {
	// Список всех символов, которые нужно экранировать в MarkdownV2
	specialChars := "_*[]()~`>#+-=|{}.!"

	var replacer strings.Builder
	for _, r := range text {
		if strings.ContainsRune(specialChars, r) {
			replacer.WriteRune('\\') // экранирующий слэш
		}
		replacer.WriteRune(r)
	}
	return replacer.String()
}

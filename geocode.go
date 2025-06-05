package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func geocodeAddress(address string) (float64, float64, error) {
	url := fmt.Sprintf(
		"https://geocode-maps.yandex.ru/1.x/?format=json&apikey=%s&geocode=%s",
		os.Getenv("API_KEY_GEOCODE"),
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

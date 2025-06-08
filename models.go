package main

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
	ReviewsBlock   string `json:"reviews"`
	MapBlock       string `json:"map"`
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

package model

import "time"

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Surname string `json:"surname"`
	Patronymic string `json:"patronymic"`
	Age  int64    `json:"age"`
	Nationality string `json:"nationality"`
	Gender string `json:"gender"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PersonCreateRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
}

type IdResponse struct{
	ID int `json:"id"`
}

type PersonUpdateRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         int64  `json:"age"`
	Nationality string `json:"nationality"`
	Gender      string `json:"gender"`
}

type PersonStats struct {
	Age         int64    `json:"age"`
	Nationality string `json:"nationality"`
	Gender      string `json:"gender"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
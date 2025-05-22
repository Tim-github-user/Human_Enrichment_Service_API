package models

type Person struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name"`                
	Surname     string  `json:"surname"`             
	Patronymic  *string `json:"patronymic,omitempty"`
	Age         *int    `json:"age,omitempty"`        
	Gender      *string `json:"gender,omitempty"`    
	Nationality *string `json:"nationality,omitempty"`
}

type PersonInput struct {
	Name       string  `json:"name" binding:"required"`   
	Surname    string  `json:"surname" binding:"required"`
	Patronymic *string `json:"patronymic,omitempty"`
}
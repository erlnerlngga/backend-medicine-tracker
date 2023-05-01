package main

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CreateNewAccountType struct {
	Name string `json:"name"`
	Email string `json:"email"`
}

type AccountType struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

type LogInType struct {
	Email string `json:"email"`
}

type ClaimsType struct {
	User_ID string `json:"user_id"`
	jwt.RegisteredClaims
}

type MedicineType struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Dosage int `json:"dosage"`
	Frequency int `json:"frequency"`
	User_ID string `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type CreateNewMedicineType struct {
	Name string `json:"name"`
	Dosage int `json:"dosage"`
	Frequency int `json:"frequency"`
	User_ID string `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type UpdateMedicineType struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Dosage int `json:"dosage"`
	Frequency int `json:"frequency"`
	Updated_at time.Time `json:"updated_at"`
}

type UpdateRequest struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Dosage int `json:"dosage"`
	Frequency int `json:"frequency"`
}

func UpdateOldMedicine(ID string, name string, dosage int, frequency int) *UpdateMedicineType {
	return &UpdateMedicineType{
		ID: ID,
		Name: name,
		Dosage: dosage,
		Frequency: frequency,
		Updated_at: time.Now().UTC(),
	}
}

type MedicineRequest struct {
	Name string `json:"name"`
	Dosage int `json:"dosage"`
	Frequency int `json:"frequency"`
	User_ID string `json:"user_id"`
}

func NewMedicine(name string, dosage int, frequency int, user_id string) *CreateNewMedicineType {
	return &CreateNewMedicineType{
		Name: name,
		Dosage: dosage,
		Frequency: frequency,
		User_ID: user_id,
		Created_at: time.Now().UTC(),
		Updated_at: time.Now().UTC(),
	}
}



package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateMedicine(med *CreateNewMedicineType) error
	GetAllMedicine(ID string) ([]*MedicineType, error)
	UpdateMedicine(med *UpdateMedicineType) error
	DeleteMedicine(ID string) error
	CreateAccount(acc *CreateNewAccountType) (*AccountType, error)
	CheckEmail(email string) (*AccountType, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := fmt.Sprintf("user=postgres dbname=medicines password=%s sslmode=disable", os.Getenv("PASSWORD"))

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("database is running...")

	return &PostgresStore{
		db: db,
	}, nil
}

// CREATE FUNCTION UUID
func (s * PostgresStore) InitFuncUUID() error {
	initFunc := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	`

	_, err := s.db.Exec(initFunc)

	return err
}

// CREATE TABLE USER
func (s *PostgresStore) CreateUserTable() error {
	createTable := `
		create table if not exists users (
			id uuid default uuid_generate_v4 (),
			name varchar(50),
			email varchar(50) unique, 
			primary key(id)
		);
	`

	_, err := s.db.Exec(createTable)

	return err
}

// CREATE TABLE MEDICINE
func (s *PostgresStore) CreateMedicineTable() error {
	createTable := `
		create table if not exists medicine (
			id uuid default uuid_generate_v4 (),
			name varchar(50),
			dosage integer,
			frequency integer,
			user_id uuid references users(id),
			created_at timestamp,
			updated_at timestamp,
			primary key (id)
		);
	`
	_, err := s.db.Exec(createTable)

	return err
}

// INIT 
func (s *PostgresStore) Init() error  {
	if err := s.InitFuncUUID(); err != nil {
		return err
	}

	if err := s.CreateUserTable(); err != nil {
		return err
	}

	return s.CreateMedicineTable()
}

// CREATE ACCOUNT
func (s *PostgresStore) CreateAccount(acc *CreateNewAccountType) (*AccountType, error) {
	newAccount := new(AccountType)

	insertQuery := `insert into users (name, email) values ($1, $2) returning *;`
	err := s.db.QueryRow(insertQuery, acc.Name, acc.Email).Scan(&newAccount.ID, &newAccount.Name, &newAccount.Email)

	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}


	return newAccount, nil
}

// CHECK EMAIL 
func (s *PostgresStore) CheckEmail(email string) (*AccountType, error) {
	account := new(AccountType)
	err := s.db.QueryRow("select * from users where email = $1;", email).Scan(&account.ID, &account.Name, &account.Email)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account %s not found", email)
	} else if err != nil {
		return nil, err
	}

	return account, nil
}

// CREATE MEDICINE
func (s *PostgresStore) CreateMedicine(med *CreateNewMedicineType) error {
	insertQuery := `insert into medicine(name, dosage, frequency, user_id, created_at, updated_at) values ($1, $2, $3, $4, $5, $6);`

	_, err := s.db.Exec(insertQuery, med.Name, med.Dosage, med.Frequency, med.User_ID, med.Created_at, med.Updated_at)

	if err != nil {
		return err
	}

	return nil
}

// GET ALL MEDICINE
func (s *PostgresStore) GetAllMedicine(ID string) ([]*MedicineType, error)  {
	rows, err := s.db.Query("select * from medicine where user_id = $1;", ID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	medicines := []*MedicineType{}
	for rows.Next() {
		m := new(MedicineType)
		if err := rows.Scan(&m.ID, &m.Name, &m.Dosage, &m.Frequency, &m.User_ID, &m.Created_at, &m.Updated_at); err != nil {
			return nil , err
		}

		medicines = append(medicines, m)
	}

	if err:= rows.Err(); err != nil {
		return nil, err
	}


	return medicines, nil
}

func (s *PostgresStore) UpdateMedicine(med *UpdateMedicineType) error {
	updateQuery := "update medicine set name = $1, dosage = $2, frequency = $3, updated_at = $4 where id = $5;"

	_, err := s.db.Exec(updateQuery, med.Name, med.Dosage, med.Frequency, med.Updated_at, med.ID)

	return err
}

func (s *PostgresStore) DeleteMedicine(ID string) error {
	deleteQuery := "delete from medicine where id = $1;"

	_, err := s.db.Exec(deleteQuery, ID)

	return err
} 
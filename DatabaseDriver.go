package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "postgres"
	dbname = "postgres"
)

type TodoItem struct {
	Id int `json:"id,omitempty"`
	Name string `json:"name"`
	Date string `json:"date"`
	Description string `json:"description"`
	UserID int `json:"userId"`
}

type PasswordType string

type User struct {
	UserID int `json:"user_id,omitempty"`
	UserName string `json:"username,omitempty"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password PasswordType `json:"password"`
}

func (PasswordType) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

type TodoDBStore interface {
	InitStore() error
	Create(todo *TodoItem) (*TodoItem, error)
	List(limit int, user_id int) ([]*TodoItem, error)
	GetOne(id int) (*TodoItem, error)
}

type UserDBStore interface {
	InitStore() error
	Create(user *User, password string) (*User, error)
	List(limit int) ([]*User, error)
	GetOne(id int) (*User, error)
	GetByUsername(username string) (*User, error)
	Validate(testPassword string, username string) (bool, error)
}

type UserStore struct {
	db * sql.DB
}

func (store *UserStore) InitStore() error {
	pSqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", pSqlInfo)
	if err != nil {
		return err
	}

	store.db = db

	return nil
}

func (store *UserStore) Create(user *User, password string) (*User, error) {
	sqlStatement := `
		INSERT INTO users (username, name, email, password, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id`
	id := 0

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}

	var createdAt = time.Now().Format("2006-01-02 15:04:05")

	err = store.db.QueryRow(sqlStatement,
		user.UserName,
		user.Name,
		user.Email,
		string(hash),
		createdAt).Scan(&id)

	if err != nil {
		return nil, err
	}

	user.UserID = id
	return user, nil
}

func (store *UserStore) GetByUsername(username string) (*User, error) {
	row := store.db.QueryRow(`SELECT user_id, username, email, name FROM users WHERE username = $1;`, username)
	var userId int
	var u string
	var email string
	var name string

	err := row.Scan(&userId, &u, &email, &name)
	if err != nil {
		return nil, err
	}

	return &User{
		Name: name,
		UserID: userId,
		UserName: username,
		Email: email,
	}, nil
}

func (store *UserStore) List(limit int) ([]*User, error) {
	panic("implement me")
}

func (store *UserStore) GetOne(id int) (*User, error) {
	row := store.db.QueryRow(`SELECT * FROM users WHERE user_id=$1`, id)
	var user *User

	err := row.Scan(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (store *UserStore) Validate(testPassword string, username string) (bool, error) {
	testBytes := []byte(testPassword)

	var password []byte

	err := store.db.QueryRow(`SELECT password from users where username = $1`,
		username).Scan(&password)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(password, testBytes)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	return true, nil
}

//// to-do item todoItemStore ////

type TodoItemStore struct {
	db * sql.DB
}

func (store *TodoItemStore) InitStore() error {
	pSqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", pSqlInfo)
	if err != nil {
		return err
	}

	store.db = db

	return nil
}

func (store *TodoItemStore) Create(todo *TodoItem) (*TodoItem, error) {
	sqlStatement := `
		INSERT INTO todoitems (name, date, description, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	id := 0
	err := store.db.QueryRow(sqlStatement, todo.Name, todo.Date, todo.Description, todo.UserID).Scan(&id)
	if err != nil {
		return nil, err
	}

	todo.Id = id
	return todo, nil
}

func (store *TodoItemStore) List(limit int, userId int) ([]*TodoItem, error) {
	rows, err := store.db.Query("SELECT name, date, description, id FROM todoitems where user_id = $1 ORDER BY date LIMIT $2", userId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todoItems []*TodoItem
	for rows.Next() {
		var name string
		var date string
		var description string
		var id int
		err = rows.Scan(&name, &date, &description, &id)
		if err != nil {
			return nil, err
		}
		item := &TodoItem{
			Name: name,
			Date: date,
			Description: description,
			Id: id,
			UserID: userId,
		}
		todoItems = append(todoItems, item)
	}

	return todoItems, nil
}

func (store *TodoItemStore) GetOne(id int) (*TodoItem, error) {
	row := store.db.QueryRow(`SELECT * FROM todoItems WHERE id=$1`, id)
	var item *TodoItem

	err := row.Scan(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}
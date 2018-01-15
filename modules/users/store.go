package users

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cjsaylor/boxmeup-go/middleware"
	jwt "github.com/dgrijalva/jwt-go"
)

// Store is a persistence structure to get and store users.
type Store struct {
	DB *sql.DB
}

// NewStore constructs a storage interface for users.
func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func hashPassword(config middleware.AuthConfig, password string) string {
	data := []byte(fmt.Sprintf("%v%v", config.LegacySalt, password))
	return fmt.Sprintf("%x", sha1.Sum(data))
}

// Login authenticates user credentials and produces a signed JWT
func (s *Store) Login(config middleware.AuthConfig, email string, password string) (string, error) {
	hashedPassword := hashPassword(config, password)
	var ID int
	var UUID string
	q := `
		select id, uuid from users where email = ? and password = ?
	`
	err := s.DB.QueryRow(q, email, hashedPassword).Scan(&ID, &UUID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        ID,
		"uuid":      UUID,
		"nbf":       time.Now().Unix(),
		"exp":       time.Now().AddDate(0, 0, 5).Unix(),
		"xsrfToken": csrfToken(),
	})
	return token.SignedString([]byte(config.JWTSecret))
}

func csrfToken() []byte {
	token := make([]byte, 256)
	rand.Read(token)
	return token
}

// Register creates a new user in the system.
// @todo Replace shitty password hashing with a more robust mechanism (bcrypt)
func (s *Store) Register(config middleware.AuthConfig, email string, password string) (id int64, err error) {
	if s.doesUserExistByEmail(email) {
		return 0, errors.New("user already exists with given email")
	}
	hashedPassword := hashPassword(config, password)
	q := `
		insert into users (email, password, uuid, created, modified)
		values (?, ?, uuid(), now(), now())
	`
	res, err := s.DB.Exec(q, email, hashedPassword)
	id, _ = res.LastInsertId()
	return
}

func (s *Store) doesUserExistByEmail(email string) bool {
	// flesh this out
	q := "select count(*) from users where email = ?"
	var count int
	s.DB.QueryRow(q, email).Scan(&count)
	return count > 0
}

// ByID resolves with a user on the channel.
func (s *Store) ByID(ID int64) (User, error) {
	user := User{}
	q := `
		select id, email, password, uuid, is_active, reset_password, created, modified
		from users where id = ?
	`
	err := s.DB.QueryRow(q, ID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.UUID,
		&user.IsActive,
		&user.ResetPassword,
		&user.Created,
		&user.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			// user not found
			// @todo consider sending a custom error that the route handler can consume
		} else {
			log.Fatal(err)
		}
	}
	return user, err
}

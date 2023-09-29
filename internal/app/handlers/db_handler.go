package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Tokebay/yandex/internal/app/storage"
	"github.com/Tokebay/yandex/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// проверяем соединение с БД
func (us *URLShortener) CheckDBConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := us.GetDB()
	if err != nil {
		logger.Log.Error("Error connect to DB", zap.Error(err))
		http.Error(w, "Error connect to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = CreateShortenedURLTable(db)
	if err != nil {
		logger.Log.Error("Error create table", zap.Error(err))
		http.Error(w, "Error create table", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CreateShortenedURLTable(db *sql.DB) error {
	// Создаем таблицу shortened_urls, если она еще не существует
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS public.shorten_urls
	(
		uuid SERIAL,
		short_url text COLLATE pg_catalog."default",
		original_url text COLLATE pg_catalog."default"
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (us *URLShortener) GetDB() (*sql.DB, error) {

	dbConnString := us.config.DSN
	fmt.Printf("dsn %s \n", dbConnString)
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		logger.Log.Error("Error open connection with DB", zap.Error(err))

	}
	return db, nil
}

// Функция для обновления идентификатора пользователя в базе данных.
func (us *URLShortener) GetNextUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	var userID int

	userID, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "internal Server Error", http.StatusInternalServerError)
		logger.Log.Error("Error getting user cookie", zap.Error(err))
		return 0, err
	}

	if userID == 0 {
		pgStorage := us.Storage.(*storage.PostgreSQLStorage)
		userID, err = pgStorage.InsertUser()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			logger.Log.Error("Error Insert Users", zap.Error(err))
			return 0, err
		}
		fmt.Printf("GetNextUserID userID %d \n", userID)
	}

	if err = SetCookieUserID(w, userID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logger.Log.Error("Error setting user ID cookie", zap.Error(err))
		return 0, fmt.Errorf("error setting user ID cookie: %w", err)
	}

	return userID, nil
}

func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	fmt.Printf("GetUserID. tokenString %s \n", tokenString)
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return -1, ErrToken
	}

	return claims.UserID, err
}

func SetCookieUserID(w http.ResponseWriter, userID int) error {
	token, err := BuildJWTString(userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Log.Error("SetCookieUserID. error BuildJWTString", zap.Error(err))
		return err
	}

	cookie := &http.Cookie{
		Name:  CookieName,
		Value: token,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (us *URLShortener) GetUserURLs(userID int) ([]URLData, error) {

	var urls []URLData
	//Prepared Statements
	pgStorage := us.Storage.(*storage.PostgreSQLStorage)
	stmt, err := pgStorage.Prepare("SELECT uuid, short_url, original_url FROM shorten_urls WHERE user_id = $1")
	if err != nil {
		logger.Log.Error("Error prepare statement", zap.Error(err))
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		logger.Log.Error("Error stmt query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url URLData
		err := rows.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
		if err != nil {
			logger.Log.Error("Error scanning rows", zap.Error(err))
			return nil, err
		}
		urls = append(urls, url)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("Error rows", zap.Error(err))
		return nil, err
	}

	return urls, nil
}

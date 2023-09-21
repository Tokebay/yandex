package handlers

import "net/http"

type User struct {
	ID string
}

type AuthService struct {
	// логика создания и проверки сессий
}

func NewAuthService() *AuthService {
	// Инициализация сервиса аутентификации
	return &AuthService{}
}

func (as *AuthService) AuthenticateUser(w http.ResponseWriter, r *http.Request) *User {
	// Если пользователь аутентифицирован, вернем объект User
	return nil
}

func (us *URLShortener) UserURLsHandler(w http.Response, r *http.Request) {

}

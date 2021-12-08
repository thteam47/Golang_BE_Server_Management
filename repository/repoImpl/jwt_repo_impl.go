package repoimpl

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
)

type JwtRepositoryImpl struct {
	secretKey     string
	tokenDuration time.Duration
}

var vi *viper.Viper

func NewJwtRepo(secretKey string, tokenDuration time.Duration) repo.JwtRepository {
	vi = viper.New()
	vi.SetConfigFile("config.yaml")
	vi.ReadInConfig()
	return &JwtRepositoryImpl{
		secretKey:     secretKey,
		tokenDuration: tokenDuration}
}
func (manager *JwtRepositoryImpl) Generate(user *models.User) (string, error) {
	claims := models.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    user.ID.Hex(),
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Role:   user.Role,
		Action: user.Action,
	}

	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := claim.SignedString([]byte(vi.GetString("keySecret")))
	if err != nil {
		return "", err
	}
	return token, nil
}
func (manager *JwtRepositoryImpl) Verify(accessToken string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&models.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil

}

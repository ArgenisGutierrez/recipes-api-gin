package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/ArgenisGutierrez/recipes-api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (handler *AuthHandler) SignInHandler(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato inválido"})
		return
	}

	// 1. Buscar solo por username primero
	var storedUser models.User
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
	}).Decode(&storedUser)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	// 2. Comparar hashes de forma segura
	hasher := sha256.New()
	hasher.Write([]byte(user.Password)) // ¡Este paso faltaba!
	inputPasswordHash := hex.EncodeToString(hasher.Sum(nil))

	if storedUser.Password != inputPasswordHash {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	sessionToken := xid.New().String()
	session := sessions.Default(ctx)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la sesión"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User signed in"})

}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "Not logged",
			})
			ctx.Abort()
		}
		ctx.Next()
	}
}

func (handler *AuthHandler) Refreshhandler(ctx *gin.Context) {
	// Obtiene el token de la cabecera
	tokenValue := ctx.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	// Verificar si recibimos un token
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Verificar si el token es valido
	if tkn == nil || !tkn.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Verificar si el token no expira al solicitar renovacion
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token not expired yet"})
		return
	}

	// Renueva el tiempo del token
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()

	//Genera un nuevo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//convierte el token a string
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	//verifica si hubo error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//crea el nuevo jwtOutput
	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}

	ctx.JSON(http.StatusOK, jwtOutput)

}

func (handler *AuthHandler) SignOutHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la sesión"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

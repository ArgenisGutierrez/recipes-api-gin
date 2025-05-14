package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	users := map[string]string{
		"admin":      "fCRmh4Q2J7Rseqkz",
		"packt":      "RE4zfHB35VPtTkbT",
		"mlabouardy": "L3nSFRcZzNQ67bcc",
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Verificar conexión
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	for username, password := range users {
		// 1. Crear nuevo hasher para cada contraseña
		h := sha256.New()

		// 2. Escribir los bytes de la contraseña en el hasher
		h.Write([]byte(password))

		// 3. Calcular hash y convertir a hexadecimal
		hash := hex.EncodeToString(h.Sum(nil))

		// 4. Insertar documento
		_, err := collection.InsertOne(ctx, bson.M{
			"username": username,
			"password": hash, // Guardar como string hexadecimal
		})

		if err != nil {
			log.Fatalf("Error insertando usuario %s: %v", username, err)
		}
	}

	log.Println("Usuarios insertados exitosamente")
}

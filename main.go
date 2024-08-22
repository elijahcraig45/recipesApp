package main

import (
	"context"
	"eTEats_backend/handlers"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	ctx := context.Background()

	// Path to your service account key file
	credentialsFilePath := "recipes-433314-92ae1fbf7aca.json"

	// Set the GOOGLE_APPLICATION_CREDENTIALS environment variable
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsFilePath)
	if err != nil {
		log.Fatalf("Failed to set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
	}

	// Initialize Firestore client
	client, err := firestore.NewClient(ctx, "recipes-433314")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Create a new router
	r := mux.NewRouter()

	// Pass the Firestore client to the handler
	r.HandleFunc("/recipes", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetRecipes(client, w, r)
	}).Methods("GET")

	// Define the route for fetching a recipe by Id
	r.HandleFunc("/recipe", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetRecipe(client, w, r)
	}).Methods("GET")

	// Define the route for creating a new recipe
	r.HandleFunc("/recipe", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateRecipe(client, w, r)
	}).Methods("POST")

	// Define the route for creating a new recipe
	r.HandleFunc("/delete/recipe", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteReceipe(client, w, r)
	}).Methods("DELETE")

	// Define the route for updating a recipe field by id
	r.HandleFunc("/update/recipe", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateRecipeField(client, w, r)
	}).Methods("PUT")

	r.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		handlers.FetchImageHandler(w, r)
	}).Methods("GET")

	// Enable CORS for all origins
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap your handlers with the CORS middleware
	handler := c.Handler(r)

	// Start the server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

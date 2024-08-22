package handlers

import (
	"context"
	"eTEats_backend/models"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"google.golang.org/api/iterator"
)

func GetRecipes(client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var recipes []models.Recipe
	iter := client.Collection("recipes").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var recipe models.Recipe
		doc.DataTo(&recipe)

		// Ensure slices are not nil
		if recipe.Ingredients == nil {
			recipe.Ingredients = []string{}
		}
		if recipe.Instructions == nil {
			recipe.Instructions = []string{}
		}
		if recipe.Tags == nil {
			recipe.Tags = []string{}
		}

		recipes = append(recipes, recipe)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recipes); err != nil {
		http.Error(w, "Failed to encode recipes", http.StatusInternalServerError)
	}
}

func GetRecipe(client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get the "id" query parameter from the URL
	recipeID := r.URL.Query().Get("id")
	if recipeID == "" {
		http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	// Query Firestore where the "id" field matches the provided recipeID
	iter := client.Collection("recipes").Where("id", "==", recipeID).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "No matching recipe found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to retrieve recipe", http.StatusInternalServerError)
		log.Printf("Failed to retrieve recipe: %v", err)
		return
	}

	// Decode the document data into the Recipe struct
	var recipe models.Recipe
	err = doc.DataTo(&recipe)
	if err != nil {
		http.Error(w, "Failed to decode recipe data", http.StatusInternalServerError)
		log.Printf("Failed to decode recipe data: %v", err)
		return
	}

	// Ensure slices are not nil
	if recipe.Ingredients == nil {
		recipe.Ingredients = []string{}
	}
	if recipe.Instructions == nil {
		recipe.Instructions = []string{}
	}
	if recipe.Tags == nil {
		recipe.Tags = []string{}
	}

	// Return the recipe as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recipe); err != nil {
		http.Error(w, "Failed to encode recipe data", http.StatusInternalServerError)
	}
}

func CreateRecipe(client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var recipe models.Recipe

	// Parse the JSON request body into the Recipe struct
	err := json.NewDecoder(r.Body).Decode(&recipe)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}

	// If the ID is not provided, generate a new one
	if recipe.ID == "" {
		recipe.ID = uuid.New().String()
	} //TODO set this to create proper id

	// Add the recipe to Firestore
	_, err = client.Collection("recipes").Doc(recipe.ID).Set(ctx, recipe)
	if err != nil {
		http.Error(w, "Failed to create recipe: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Failed to create recipe: %v", err)
		return
	}

	// Return the created recipe as a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(recipe); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func DeleteReceipe(client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get the "id" query parameter from the URL
	recipeID := r.URL.Query().Get("id")
	if recipeID == "" {
		http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	// Delete the document from Firestore
	_, err := client.Collection("recipes").Doc(recipeID).Delete(ctx)
	if err != nil {
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		log.Printf("Failed to delete recipe with id %s: %v", recipeID, err)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusNoContent) // 204 No Content indicates successful deletion
}

type UpdateFieldRequest struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

func UpdateRecipeField(client *firestore.Client, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get the "id" query parameter from the URL
	recipeID := r.URL.Query().Get("id")
	if recipeID == "" {
		http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	// Parse the JSON body to get the field name and new value
	var updateRequest UpdateFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}

	// Prepare the update data
	updateData := map[string]interface{}{
		updateRequest.Field: updateRequest.Value,
	}

	// Update the document in Firestore
	_, err := client.Collection("recipes").Doc(recipeID).Set(ctx, updateData, firestore.MergeAll)
	if err != nil {
		http.Error(w, "Failed to update recipe field", http.StatusInternalServerError)
		log.Printf("Failed to update recipe with id %s: %v", recipeID, err)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recipe field updated successfully"))
}

// FetchImageHandler fetches an image from a URL, resizes it, and returns it.
func FetchImageHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight (OPTIONS) request
	if r.Method == http.MethodOptions {
		return
	}

	// Get the URL parameter
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	// Fetch the image from the URL
	resp, err := http.Get(imageURL)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	// Calculate new width while maintaining aspect ratio
	newHeight := uint(500)
	originalBounds := img.Bounds()
	aspectRatio := float64(originalBounds.Dx()) / float64(originalBounds.Dy())
	newWidth := uint(float64(newHeight) * aspectRatio)

	// Resize the image
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// Set the content type
	w.Header().Set("Content-Type", "image/"+format)

	// Encode the resized image and write it to the response
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(w, resizedImg, nil)
	case "png":
		err = png.Encode(w, resizedImg)
	default:
		http.Error(w, "Unsupported image format", http.StatusUnsupportedMediaType)
		return
	}

	if err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
	}
}

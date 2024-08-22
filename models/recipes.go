package models

type Recipe struct {
	ID           string   `firestore:"id"`
	Name         string   `firestore:"Name"`
	Description  string   `firestore:"Description"`
	Ingredients  []string `firestore:"Ingredients"`
	Instructions []string `firestore:"Instructions"`
	Notes        string   `firestore:"Notes"`
	Tags         []string `firestore:"tags"`
	ImageURL     string   `firestore:"imageURL"`
	OriginalURL  string   `firestore:"OriginalURL"`
}

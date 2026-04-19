package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Vendor string `json:"vendor"`
	Price int `json:"price"`
	ImageURL string `json:"image_url"`
}

func main() {
	db, err := sql.Open("sqlite", "file:shop.db?cache=shared&mode=rwc")
	if err!= nil { log.Fatal(err) }

	db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY, name TEXT, vendor TEXT, price INTEGER, image_url TEXT
	)`)

	// পুরনো ডাটা মুছে নতুন ছবিসহ ডাটা
	db.Exec(`DELETE FROM products`)
	db.Exec(`INSERT INTO products (name, vendor, price, image_url) VALUES
		('Wireless Headphone','TechStore BD',2999,'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600'),
		('Panjabi Premium','Deshi Fashion',1890,'https://images.unsplash.com/photo-1596755094514-f87e34085b2c?w=600'),
		('Arabian Dates 1kg','Qatar Mart',1200,'https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=600')
	`)

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		rows, _ := db.Query("SELECT id, name, vendor, price, image_url FROM products")
		defer rows.Close()
		var products []Product
		for rows.Next() {
			var p Product
			rows.Scan(&p.ID, &p.Name, &p.Vendor, &p.Price, &p.ImageURL)
			products = append(products, p)
		}
		json.NewEncoder(w).Encode(products)
	})

	http.HandleFunc("/cart/add", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { return }
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(`[]`))
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Println("API running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

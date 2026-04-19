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

	db.Exec(`CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY, name TEXT, vendor TEXT, price INTEGER, image_url TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY, items TEXT, total INTEGER, payment_method TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if count == 0 {
		db.Exec(`INSERT INTO products (name, vendor, price, image_url) VALUES
			('Wireless Headphone','TechStore BD',2999,'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600'),
			('Panjabi Premium','Deshi Fashion',1890,'https://images.unsplash.com/photo-1596755094514-f87e34085b2c?w=600'),
			('Arabian Dates 1kg','Qatar Mart',1200,'https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=600')
		`)
	}

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

	http.HandleFunc("/admin/add-product", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { return }
		var req struct {
			Name string; Vendor string; Price int; ImageURL string `json:"image_url"`; Secret string
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Secret!= "haymi2026" { http.Error(w, "unauthorized", 401); return }
		db.Exec("INSERT INTO products (name, vendor, price, image_url) VALUES (?,?,?,?)", req.Name, req.Vendor, req.Price, req.ImageURL)
		w.Write([]byte(`{"status":"added"}`))
	})

	// NEW: Create Order
	http.HandleFunc("/order/create", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { return }
		
		var req struct {
			Items string `json:"items"`
			Total int `json:"total"`
			PaymentMethod string `json:"payment_method"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		res, _ := db.Exec("INSERT INTO orders (items, total, payment_method) VALUES (?,?,?)", req.Items, req.Total, req.PaymentMethod)
		id, _ := res.LastInsertId()
		json.NewEncoder(w).Encode(map[string]any{"order_id": id, "status": "created"})
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

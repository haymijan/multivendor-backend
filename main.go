package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Variant struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	SKU       string   `json:"sku"`
	Color     string   `json:"color"`
	Size      string   `json:"size"`
	Price     float64  `json:"price"`
	Stock     int      `json:"stock"`
	Images    []string `json:"images"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	
	// Create variants table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS variants (
		id SERIAL PRIMARY KEY,
		product_id INT NOT NULL,
		sku TEXT UNIQUE NOT NULL,
		color TEXT,
		size TEXT,
		price NUMERIC(10,2) NOT NULL,
		stock INT DEFAULT 0,
		images TEXT[],
		created_at TIMESTAMP DEFAULT NOW()
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/variants", variantsHandler)
	http.HandleFunc("/api/variants/", variantByIDHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Variants backend running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func variantsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		productID := r.URL.Query().Get("product_id")
		rows, err := db.Query("SELECT id, product_id, sku, color, size, price, stock, images FROM variants WHERE product_id=$1 ORDER BY id", productID)
		if err != nil {
			http.Error(w, err.Error(), 500); return
		}
		defer rows.Close()
		var variants []Variant
		for rows.Next() {
			var v Variant
			var images []string
			err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Color, &v.Size, &v.Price, &v.Stock, &images)
			if err == nil {
				v.Images = images
				variants = append(variants, v)
			}
		}
		json.NewEncoder(w).Encode(variants)
		
	case "POST":
		var v Variant
		json.NewDecoder(r.Body).Decode(&v)
		v.SKU = strings.ToUpper(strings.ReplaceAll(v.SKU, " ", "-"))
		err := db.QueryRow(`INSERT INTO variants (product_id, sku, color, size, price, stock, images) 
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
			v.ProductID, v.SKU, v.Color, v.Size, v.Price, v.Stock, v.Images).Scan(&v.ID)
		if err != nil {
			http.Error(w, err.Error(), 500); return
		}
		json.NewEncoder(w).Encode(v)
		
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func variantByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	idStr := strings.TrimPrefix(r.URL.Path, "/api/variants/")
	id, _ := strconv.Atoi(idStr)
	
	if r.Method == "PUT" {
		var v Variant
		json.NewDecoder(r.Body).Decode(&v)
		_, err := db.Exec(`UPDATE variants SET color=$1, size=$2, price=$3, stock=$4, images=$5 WHERE id=$6`,
			v.Color, v.Size, v.Price, v.Stock, v.Images, id)
		if err != nil {
			http.Error(w, err.Error(), 500); return
		}
		json.NewEncoder(w).Encode(map[string]bool{"updated": true})
	} else if r.Method == "DELETE" {
		db.Exec("DELETE FROM variants WHERE id=$1", id)
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

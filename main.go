package main

import (
  "database/sql"
  "encoding/json"
  "log"
  "net/http"
  _ "modernc.org/sqlite"
)

type Product struct { ID int `json:"id"`; Name string `json:"name"`; Price int `json:"price"`; Vendor string `json:"vendor"` }
type CartItem struct { Product Product `json:"product"`; Qty int `json:"qty"` }

var db *sql.DB

func withCORS(h http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    if r.Method=="OPTIONS"{w.WriteHeader(200);return}
    h(w,r)
  }
}

func main(){
  db,_ = sql.Open("sqlite","./shop.db")
  db.Exec(`CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price INTEGER, vendor TEXT)`)
  db.Exec(`CREATE TABLE IF NOT EXISTS cart (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id INTEGER, qty INTEGER)`)

  var c int; db.QueryRow("SELECT COUNT(*) FROM products").Scan(&c)
  if c==0 {
    db.Exec("INSERT INTO products(name,price,vendor) VALUES('Wireless Headphone',2999,'TechStore BD')")
    db.Exec("INSERT INTO products(name,price,vendor) VALUES('Panjabi Premium',1890,'Deshi Fashion')")
    db.Exec("INSERT INTO products(name,price,vendor) VALUES('Arabian Dates 1kg',1200,'Qatar Mart')")
  }

  http.HandleFunc("/products", withCORS(func(w http.ResponseWriter, r *http.Request){
    if r.Method=="POST" {
      var p Product; json.NewDecoder(r.Body).Decode(&p)
      db.Exec("INSERT INTO products(name,price,vendor) VALUES(?,?,?)", p.Name, p.Price, p.Vendor)
      w.WriteHeader(201); json.NewEncoder(w).Encode(map[string]string{"status":"created"}); return
    }
    rows,_ := db.Query("SELECT id,name,price,vendor FROM products"); defer rows.Close()
    list:=[]Product{}; for rows.Next(){var p Product; rows.Scan(&p.ID,&p.Name,&p.Price,&p.Vendor); list=append(list,p)}
    json.NewEncoder(w).Encode(list)
  }))

  http.HandleFunc("/cart", withCORS(func(w http.ResponseWriter, r *http.Request){
    rows,_ := db.Query(`SELECT p.id,p.name,p.price,p.vendor,c.qty FROM cart c JOIN products p ON p.id=c.product_id`); defer rows.Close()
    list:=[]CartItem{}; for rows.Next(){var ci CartItem; rows.Scan(&ci.Product.ID,&ci.Product.Name,&ci.Product.Price,&ci.Product.Vendor,&ci.Qty); list=append(list,ci)}
    json.NewEncoder(w).Encode(list)
  }))
  http.HandleFunc("/cart/add", withCORS(func(w http.ResponseWriter, r *http.Request){
    var req struct{ID int `json:"id"`}; json.NewDecoder(r.Body).Decode(&req)
    db.Exec("INSERT INTO cart(product_id,qty) VALUES(?,1)", req.ID)
    json.NewEncoder(w).Encode(map[string]string{"status":"added"})
  }))
  http.HandleFunc("/cart/clear", withCORS(func(w http.ResponseWriter, r *http.Request){
    db.Exec("DELETE FROM cart")
    json.NewEncoder(w).Encode(map[string]string{"status":"cleared"})
  }))

  log.Println("API running on http://localhost:8080"); log.Fatal(http.ListenAndServe(":8080",nil))
}
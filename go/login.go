package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
	"encoding/json"
    "github.com/joho/godotenv"
    "github.com/gorilla/sessions"
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
    "github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func setupDB() *sql.DB {
    db, err := sql.Open("mysql", "root:@tcp(localhost)/smoobites")
    if err != nil {
        log.Fatal(err)
    }
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
    return db
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func authHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing the form", http.StatusBadRequest)
            return
        }

        action := r.FormValue("action")
        email := r.FormValue("email")
        password := r.FormValue("password")

        if action == "register" {
            // Register user
            name := r.FormValue("name")

            hashedPassword, err := hashPassword(password)
            if err != nil {
                log.Printf("Failed to hash password: %v", err)
                http.Error(w, "Unable to hash password", http.StatusBadRequest)
                return
            }

            role := "user" // Default role, can be changed based on requirements

            _, err = db.Exec("INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)", name, email, hashedPassword, role)
            if err != nil {
                log.Printf("Failed to insert user: %v", err)
                http.Error(w, "Unable to insert user", http.StatusBadRequest)
                return
            }

            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            fmt.Fprintf(w, "<h1>Registration successful!</h1>")
        } else if action == "login" {
            // Login user
            var storedPassword, role, name string
            var vendorId int
            err := db.QueryRow("SELECT password, role, name, id FROM users WHERE email = ?", email).Scan(&storedPassword, &role, &name, &vendorId)
            if err != nil {
                if err == sql.ErrNoRows {
                    http.Error(w, "Unable to find user", http.StatusBadRequest)
                } else {
                    log.Printf("Failed to query user: %v", err)
                    http.Error(w, "Unable to find user", http.StatusBadRequest)
                }
                return
            }

            if !checkPasswordHash(password, storedPassword) {
                http.Error(w, "Unable to find user", http.StatusBadRequest)
                return
            }

            // Create session
            session, _ := store.Get(r, "session-name")
            session.Values["user"] = name
            session.Values["role"] = role
            session.Values["vendorId"] = vendorId
            session.Save(r, w)

            if role == "vendor" {
                http.Redirect(w, r, "/vmainpage.html", http.StatusSeeOther)
            } else {
                http.Redirect(w, r, "/index.html", http.StatusSeeOther)
            }

        } else {
            http.Error(w, "Invalid action", http.StatusBadRequest)
        }
    }
}


func sessionInfoHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")
        user := session.Values["user"]
        role := session.Values["role"]
        vendorId := session.Values["vendorId"]

        w.Header().Set("Content-Type", "application/json")
        if user != nil && role != nil && vendorId != nil {
            fmt.Fprintf(w, `{"user":"%s", "role":"%s", "vendorId":%d}`, user, role, vendorId)
        } else {
            fmt.Fprintf(w, `{"user":null, "role":null, "vendorId":null}`)
        }
    }
}

func logoutHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")
        session.Options.MaxAge = -1 // This will clear the session
        session.Save(r, w)
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"status":"logged out"}`)
    }
}



func createCheckoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var cart struct {
		Items []struct {
			Name     string  `json:"name"`
			Quantity int64   `json:"quantity"`
			Price    float64 `json:"price"`
		} `json:"cart"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams
	for _, item := range cart.Items {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(string(stripe.CurrencyUSD)),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name),
				},
				UnitAmount: stripe.Int64(int64(item.Price * 100)),
			},
			Quantity: stripe.Int64(item.Quantity),
		})
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String("https://your-domain.com/success"),
		CancelURL:          stripe.String("https://your-domain.com/cancel"),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("Session creation failed: %v\n", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": s.ID,
	})
}
func main() {
    db := setupDB()
    defer db.Close()
    err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Set  Stripe secret key from the environment variable
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

    // Serve index.html for the root URL
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.ServeFile(w, r, filepath.Join("..", "index.html"))
        } else {
            filePath := filepath.Join("..", r.URL.Path[1:]) // Remove the leading slash
            if _, err := os.Stat(filePath); os.IsNotExist(err) {
                http.NotFound(w, r)
                return
            }
            http.ServeFile(w, r, filePath)
        }
    })

    // API endpoints
    http.HandleFunc("/auth", authHandler(db))
    http.HandleFunc("/session-info", sessionInfoHandler())
    http.HandleFunc("/logout", logoutHandler())
    http.HandleFunc("/add-food-item", addFoodItemHandler(db))
    http.HandleFunc("/get-food-items", getFoodItemsHandler(db)) 
    http.HandleFunc("/get-food-details", getFoodDetailsHandler(db)) 
    http.HandleFunc("/update-food-item", updateFoodItemHandler(db))
    http.HandleFunc("/get-food-item", getFoodItemByIDHandler(db))
    http.HandleFunc("/set-food-session", setFoodSessionHandler())
    http.HandleFunc("/delete-food-item", deleteFoodItemHandler(db))
    http.HandleFunc("/create-checkout-session", createCheckoutSession)
    http.HandleFunc("/forgot-password", forgotPasswordHandler(db))
	http.HandleFunc("/reset-password", resetPasswordHandler(db))


    log.Println("Server started on http://localhost:5500")
    if err := http.ListenAndServe(":5500", nil); err != nil {
        log.Fatal(err)
    }
}

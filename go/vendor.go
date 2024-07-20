package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
	"fmt"

    _ "github.com/go-sql-driver/mysql"
    
)

type Response struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}




func addFoodItemHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        log.Println("Parsing form data...")
        err := r.ParseMultipartForm(10 << 20) // 10 MB
        if err != nil {
            log.Printf("Error parsing form: %v\n", err)
            http.Error(w, "Unable to parse form", http.StatusBadRequest)
            return
        }

        foodName := r.FormValue("foodName")
        description := r.FormValue("description")
        price := r.FormValue("price")
        prepTime := r.FormValue("prepTime")

        // Retrieve vendor ID from the session
        session, err := store.Get(r, "session-name")
        if err != nil {
            log.Printf("Error getting session: %v\n", err)
            http.Error(w, "Error getting session", http.StatusInternalServerError)
            return
        }
        vendorID, ok := session.Values["vendorId"].(int)
        if !ok {
            log.Printf("Invalid vendor ID in session")
            http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
            return
        }

        log.Printf("Retrieved form data: foodName=%s, description=%s, price=%s, prepTime=%s, vendorID=%d",
            foodName, description, price, prepTime, vendorID)

        log.Println("Retrieving file...")
        file, handler, err := r.FormFile("foodImage")
        if err != nil {
            log.Printf("Error retrieving the file: %v\n", err)
            http.Error(w, "Error retrieving the file", http.StatusBadRequest)
            return
        }
        defer file.Close()

        log.Printf("Saving file: %s\n", handler.Filename)
        // Use filepath.Join to create a cross-platform compatible file path
        filePath := filepath.Join("uploads", handler.Filename)
        // Ensure the 'uploads' directory exists
        if _, err := os.Stat("uploads"); os.IsNotExist(err) {
            err = os.Mkdir("uploads", 0755)
            if err != nil {
                log.Printf("Unable to create uploads directory: %v\n", err)
                http.Error(w, "Unable to create uploads directory", http.StatusInternalServerError)
                return
            }
        }
        f, err := os.Create(filePath)
        if err != nil {
            log.Printf("Unable to save the file: %v\n", err)
            http.Error(w, "Unable to save the file", http.StatusInternalServerError)
            return
        }
        defer f.Close()
        _, err = f.ReadFrom(file)
        if err != nil {
            log.Printf("Unable to save the file: %v\n", err)
            http.Error(w, "Unable to save the file", http.StatusInternalServerError)
            return
        }

        log.Println("Inserting data into the database...")
        res, err := db.Exec("INSERT INTO food_items (food_name, description, price, prep_time, image_path, vendor_id) VALUES (?, ?, ?, ?, ?, ?)",
            foodName, description, price, prepTime, filePath, vendorID)
        if err != nil {
            log.Printf("Error inserting data: %v\n", err)
            response := Response{Success: false, Message: "Unable to save the data"}
            json.NewEncoder(w).Encode(response)
            return
        }

        foodID, err := res.LastInsertId()
        if err != nil {
            log.Printf("Error retrieving last insert ID: %v\n", err)
            response := Response{Success: false, Message: "Unable to retrieve last insert ID"}
            json.NewEncoder(w).Encode(response)
            return
        }

        // Iterate through form values to retrieve add-ons
        for i := 1; i <= 10; i++ {
            addOnName := r.FormValue(fmt.Sprintf("addonName%d", i))
            addOnPrice := r.FormValue(fmt.Sprintf("addonPrice%d", i))
            if addOnName != "" && addOnPrice != "" {
                _, err := db.Exec("INSERT INTO addons (food_id, name, price) VALUES (?, ?, ?)", foodID, addOnName, addOnPrice)
                if err != nil {
                    log.Printf("Error inserting add-on: %v\n", err)
                    response := Response{Success: false, Message: "Unable to save add-ons"}
                    json.NewEncoder(w).Encode(response)
                    return
                }
            }
        }

        response := Response{Success: true, Message: "Food item added successfully"}
        json.NewEncoder(w).Encode(response)
    }
}

// Addon represents an add-on for a food item
type Addon struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

type FoodItem struct {
    ID          int     `json:"id"`
    FoodName    string  `json:"food_name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    PrepTime    int     `json:"prep_time"`
    ImagePath   string  `json:"image_path"`
    VendorID    int     `json:"vendor_id"`
    Addons      []Addon `json:"addons"` // Updated to []Addon
}
type Vendor struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	FoodItems []FoodItem `json:"food_items"`
}

type Order struct {
    UserID    int     `json:"user_id"`
    VendorID  int     `json:"vendor_id"`
    TotalPrice float64 `json:"total_price"`
    Status    string  `json:"status"`
}

type OrderItem struct {
    OrderID   int     `json:"order_id"`
    FoodID    int     `json:"food_id"`
    Quantity  int     `json:"quantity"`
    AddonsIDs []int   `json:"addons_ids"`
    ItemPrice float64 `json:"item_price"`
}

type CartItem struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    Price   float64 `json:"price"`
    Quantity int    `json:"quantity"`
    Addons  []Addon `json:"addons"`
}
func getFoodItemsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("Received request to fetch food items.")
        session, err := store.Get(r, "session-name")
        log.Println("Session Values:")
        for key, value := range session.Values {
            log.Printf("%v: %v\n", key, value)
        }
        if err != nil {
            log.Printf("Error retrieving session: %v\n", err)
            http.Error(w, "Session error", http.StatusInternalServerError)
            return
        }
        vendorID, ok := session.Values["vendorId"].(int)
        if !ok {
            log.Println("VendorID not found in session")
            http.Error(w, "Session error", http.StatusInternalServerError)
            return
        }

        rows, err := db.Query("SELECT id, food_name, description, price, prep_time, image_path, vendor_id FROM food_items WHERE vendor_id = ?", vendorID)
        if err != nil {
            log.Printf("Error querying food items: %v\n", err)
            http.Error(w, "Unable to retrieve food items", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var foodItems []FoodItem
        for rows.Next() {
            var item FoodItem
            var priceStr, prepTimeStr string
            var description *string 

            if err := rows.Scan(&item.ID, &item.FoodName, &item.Description, &priceStr, &prepTimeStr, &item.ImagePath, &item.VendorID); err != nil {
                log.Printf("Error scanning food item: %v\n", err)
                http.Error(w, "Unable to retrieve food items", http.StatusInternalServerError)
                return
            }

            // Check if priceStr is empty
            if priceStr != "" {
                item.Price, err = strconv.ParseFloat(priceStr, 64)
                if err != nil {
                    log.Printf("Error converting price to float64: %v\n", err)
                    http.Error(w, "Unable to retrieve food items", http.StatusInternalServerError)
                    return
                }
            } else {
                item.Price = 0.0 // Default value if price is empty
            }

            // Check if prepTimeStr is empty
            if prepTimeStr != "" {
                item.PrepTime, err = strconv.Atoi(prepTimeStr)
                if err != nil {
                    log.Printf("Error converting prep time to int: %v\n", err)
                    http.Error(w, "Unable to retrieve food items", http.StatusInternalServerError)
                    return
                }
            } else {
                item.PrepTime = 0 // Default value if prep time is empty
            }

            if description != nil {
                item.Description = *description
            } else {
                item.Description = ""  // Default if NULL
            }
            foodItems = append(foodItems, item)
        }

        if err := rows.Err(); err != nil {
            log.Printf("Error with rows: %v\n", err)
            http.Error(w, "Unable to retrieve food items", http.StatusInternalServerError)
            return
        }

        log.Printf("Successfully retrieved %d food items\n", len(foodItems))

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(foodItems); err != nil {
            log.Printf("Error encoding food items to JSON: %v\n", err)
            http.Error(w, "Unable to encode food items", http.StatusInternalServerError)
            return
        }

        log.Println("Response sent successfully.")
    }
}




func getFoodDetailsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")
        vendorID := session.Values["vendorid"].(int)
        foodID := session.Values["foodid"].(int)

        var foodItem FoodItem
        var addons []Addon

        // Query the food item
        var priceStr, prepTimeStr string
        err := db.QueryRow("SELECT id, food_name, description, price, prep_time, image_path, vendor_id FROM food_items WHERE vendor_id = ? AND id = ?", vendorID, foodID).
            Scan(&foodItem.ID, &foodItem.FoodName, &foodItem.Description, &priceStr, &prepTimeStr, &foodItem.ImagePath, &foodItem.VendorID)
        if err != nil {
            log.Printf("Error querying food item: %v\n", err)
            http.Error(w, "Unable to retrieve food item", http.StatusInternalServerError)
            return
        }

        // Convert price and prep time
        if priceStr != "" {
            foodItem.Price, err = strconv.ParseFloat(priceStr, 64)
            if err != nil {
                log.Printf("Error converting price to float64: %v\n", err)
                http.Error(w, "Invalid price format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.Price = 0.0 
        }

        if prepTimeStr != "" {
            foodItem.PrepTime, err = strconv.Atoi(prepTimeStr)
            if err != nil {
                log.Printf("Error converting prep time to int: %v\n", err)
                http.Error(w, "Invalid prep time format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.PrepTime = 0 
        }

        // Query the add-ons
        rows, err := db.Query("SELECT name, price FROM addons WHERE food_id = ?", foodID)
        if err != nil {
            log.Printf("Error querying addons: %v\n", err)
            http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        for rows.Next() {
            var addon Addon
            var priceStr string
            if err := rows.Scan(&addon.Name, &priceStr); err != nil {
                log.Printf("Error scanning addon: %v\n", err)
                http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
                return
            }
            addon.Price, err = strconv.ParseFloat(priceStr, 64)
            if err != nil {
                log.Printf("Error converting addon price to float64: %v\n", err)
                http.Error(w, "Invalid addon price format", http.StatusInternalServerError)
                return
            }
            addons = append(addons, addon)
        }

        if err := rows.Err(); err != nil {
            log.Printf("Error with rows: %v\n", err)
            http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
            return
        }

        foodItem.Addons = addons 

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(foodItem); err != nil {
            log.Printf("Error encoding food item to JSON: %v\n", err)
            http.Error(w, "Unable to encode food item", http.StatusInternalServerError)
            return
        }
    }
}

func getFoodItemByIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")

        // Retrieve the session values as integers
        foodID, ok := session.Values["foodid"].(int)
        if !ok {
            log.Printf("Error converting foodid to int: %v\n", session.Values["foodid"])
            http.Error(w, "Invalid food ID", http.StatusBadRequest)
            return
        }

        vendorID, ok := session.Values["vendorid"].(int)
        if !ok {
            log.Printf("Error converting vendorid to int: %v\n", session.Values["vendorid"])
            http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
            return
        }

        var foodItem FoodItem
        var addons []Addon

        // Query the food item
        var priceStr, prepTimeStr string
        err := db.QueryRow("SELECT id, food_name, description, price, prep_time, image_path, vendor_id FROM food_items WHERE vendor_id = ? AND id = ?", vendorID, foodID).
            Scan(&foodItem.ID, &foodItem.FoodName, &foodItem.Description, &priceStr, &prepTimeStr, &foodItem.ImagePath, &foodItem.VendorID)
        if err != nil {
            log.Printf("Error querying food item: %v\n", err)
            http.Error(w, "Unable to retrieve food item", http.StatusInternalServerError)
            return
        }

        // Convert price and prep time
        if priceStr != "" {
            foodItem.Price, err = strconv.ParseFloat(priceStr, 64)
            if err != nil {
                log.Printf("Error converting price to float64: %v\n", err)
                http.Error(w, "Invalid price format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.Price = 0.0 
        }

        if prepTimeStr != "" {
            foodItem.PrepTime, err = strconv.Atoi(prepTimeStr)
            if err != nil {
                log.Printf("Error converting prep time to int: %v\n", err)
                http.Error(w, "Invalid prep time format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.PrepTime = 0 
        }

        // Query the add-ons
        rows, err := db.Query("SELECT name, price FROM addons WHERE food_id = ?", foodID)
        if err != nil {
            log.Printf("Error querying addons: %v\n", err)
            http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        for rows.Next() {
            var addon Addon
            var priceStr string
            if err := rows.Scan(&addon.Name, &priceStr); err != nil {
                log.Printf("Error scanning addon: %v\n", err)
                http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
                return
            }
            addon.Price, err = strconv.ParseFloat(priceStr, 64)
            if err != nil {
                log.Printf("Error converting addon price to float64: %v\n", err)
                http.Error(w, "Invalid addon price format", http.StatusInternalServerError)
                return
            }
            addons = append(addons, addon)
        }

        if err := rows.Err(); err != nil {
            log.Printf("Error with rows: %v\n", err)
            http.Error(w, "Unable to retrieve add-ons", http.StatusInternalServerError)
            return
        }

        foodItem.Addons = addons

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(foodItem); err != nil {
            log.Printf("Error encoding food item to JSON: %v\n", err)
            http.Error(w, "Unable to encode food item", http.StatusInternalServerError)
            return
        }
    }
}

func updateFoodItemHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        log.Println("Parsing form data...")
        err := r.ParseMultipartForm(10 << 20) // 10 MB
        if err != nil {
            log.Printf("Error parsing form: %v\n", err)
            http.Error(w, "Unable to parse form", http.StatusBadRequest)
            return
        }

        // Retrieve foodID and vendorID from the session
        session, _ := store.Get(r, "session-name")

        foodID, ok := session.Values["foodid"].(int)
        if !ok {
            log.Printf("Error retrieving foodid from session: %v\n", session.Values["foodid"])
            http.Error(w, "Invalid food ID in session", http.StatusBadRequest)
            return
        }

        vendorID, ok := session.Values["vendorid"].(int)
        if !ok {
            log.Printf("Error retrieving vendorid from session: %v\n", session.Values["vendorid"])
            http.Error(w, "Invalid vendor ID in session", http.StatusBadRequest)
            return
        }

        foodName := r.FormValue("foodName")
        description := r.FormValue("description")
        price := r.FormValue("price")
        prepTime := r.FormValue("prepTime")

        log.Printf("Retrieved form data: foodName=%s, description=%s, price=%s, prepTime=%s",
            foodName, description, price, prepTime)

        var filePath string
        file, handler, err := r.FormFile("foodImage")
        if err != nil {
            log.Printf("No file uploaded: %v\n", err)
        } else {
            defer file.Close()
            log.Printf("Saving file: %s\n", handler.Filename)
            filePath = filepath.Join("uploads", handler.Filename)
            if _, err := os.Stat("uploads"); os.IsNotExist(err) {
                err = os.Mkdir("uploads", 0755)
                if err != nil {
                    log.Printf("Unable to create uploads directory: %v\n", err)
                    http.Error(w, "Unable to create uploads directory", http.StatusInternalServerError)
                    return
                }
            }
            f, err := os.Create(filePath)
            if err != nil {
                log.Printf("Unable to save the file: %v\n", err)
                http.Error(w, "Unable to save the file", http.StatusInternalServerError)
                return
            }
            defer f.Close()
            _, err = f.ReadFrom(file)
            if err != nil {
                log.Printf("Unable to save the file: %v\n", err)
                http.Error(w, "Unable to save the file", http.StatusInternalServerError)
                return
            }
        }

        log.Println("Updating data in the database...")
        query := "UPDATE food_items SET food_name = ?, description = ?, price = ?, prep_time = ?"
        args := []interface{}{foodName, description, price, prepTime}
        if filePath != "" {
            query += ", image_path = ?"
            args = append(args, filePath)
        }
        query += " WHERE id = ? and vendor_id = ?"
        args = append(args, foodID, vendorID)

        _, err = db.Exec(query, args...)
        if err != nil {
            log.Printf("Error updating data: %v\n", err)
            response := Response{Success: false, Message: "Unable to update the data"}
            json.NewEncoder(w).Encode(response)
            return
        }

        _, err = db.Exec("DELETE FROM addons WHERE food_id = ?", foodID)
        if err != nil {
            log.Printf("Error deleting existing add-ons: %v\n", err)
            response := Response{Success: false, Message: "Unable to update add-ons"}
            json.NewEncoder(w).Encode(response)
            return
        }

        for i := 1; i <= 20; i++ {
            addOnName := r.FormValue(fmt.Sprintf("addonName%d", i))
            addOnPrice := r.FormValue(fmt.Sprintf("addonPrice%d", i))
            if addOnName != "" && addOnPrice != "" {
                _, err := db.Exec("INSERT INTO addons (food_id, name, price) VALUES (?, ?, ?)", foodID, addOnName, addOnPrice)
                if err != nil {
                    log.Printf("Error inserting add-on: %v\n", err)
                    response := Response{Success: false, Message: "Unable to update add-ons"}
                    json.NewEncoder(w).Encode(response)
                    return
                }
            }
        }

        response := Response{Success: true, Message: "Food item updated successfully"}
        json.NewEncoder(w).Encode(response)
    }
}

func setFoodSessionHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session-name")

        foodIDStr := r.FormValue("foodid")
        vendorIDStr := r.FormValue("vendorid")

        // Convert strings to integers
        foodID, err := strconv.Atoi(foodIDStr)
        if err != nil {
            log.Printf("Error converting foodid to int: %v\n", err)
            http.Error(w, "Invalid food ID", http.StatusBadRequest)
            return
        }

        vendorID, err := strconv.Atoi(vendorIDStr)
        if err != nil {
            log.Printf("Error converting vendorid to int: %v\n", err)
            http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
            return
        }

        session.Values["foodid"] = foodID
        session.Values["vendorid"] = vendorID
        session.Save(r, w)

        response := map[string]string{"status": "success"}
        json.NewEncoder(w).Encode(response)
    }
}

func deleteFoodItemHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var input struct {
            FoodID   int `json:"foodid"`
            VendorID int `json:"vendorid"`
        }

        err := json.NewDecoder(r.Body).Decode(&input)
        if err != nil {
            log.Printf("Error decoding request body: %v\n", err)
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        tx, err := db.Begin()
        if err != nil {
            log.Printf("Error starting transaction: %v\n", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        defer tx.Rollback()

        // Delete add-ons associated with the food item
        _, err = tx.Exec("DELETE FROM addons WHERE food_id = ?", input.FoodID)
        if err != nil {
            log.Printf("Error deleting add-ons: %v\n", err)
            http.Error(w, "Failed to delete add-ons", http.StatusInternalServerError)
            return
        }

        // Delete the food item
        _, err = tx.Exec("DELETE FROM food_items WHERE id = ? AND vendor_id = ?", input.FoodID, input.VendorID)
        if err != nil {
            log.Printf("Error deleting food item: %v\n", err)
            http.Error(w, "Failed to delete food item", http.StatusInternalServerError)
            return
        }

        err = tx.Commit()
        if err != nil {
            log.Printf("Error committing transaction: %v\n", err)
            http.Error(w, "Failed to delete food item", http.StatusInternalServerError)
            return
        }

        response := Response{Success: true, Message: "Food item deleted successfully"}
        json.NewEncoder(w).Encode(response)
    }
}





func getVendorsBySchoolHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        school := r.URL.Query().Get("school")
        if school == "" {
            http.Error(w, "School parameter is required", http.StatusBadRequest)
            return
        }

        log.Println("Received request for school:", school)

        query := `
            SELECT u.id, u.name, f.id, f.food_name, f.description, f.price, f.image_path, 
                   COALESCE(a.id, 0) AS addon_id, 
                   COALESCE(a.name, '') AS addon_name, 
                   COALESCE(a.price, 0.0) AS addon_price
            FROM users u
            JOIN food_items f ON u.id = f.vendor_id
            LEFT JOIN addons a ON f.id = a.food_id
            WHERE u.school = ?`

        log.Println("Executing query:", query, "with school:", school)

        rows, err := db.Query(query, school)
        if err != nil {
            http.Error(w, "Failed to query database", http.StatusInternalServerError)
            log.Println("Failed to query database:", err)
            return
        }
        defer rows.Close()

        vendors := make(map[int]*Vendor)
        for rows.Next() {
            var vendorId int
            var vendorName string
            var food FoodItem
            var addon Addon

            err = rows.Scan(&vendorId, &vendorName, &food.ID, &food.FoodName, &food.Description, &food.Price, &food.ImagePath, &addon.ID, &addon.Name, &addon.Price)
            if err != nil {
                http.Error(w, "Failed to read database results", http.StatusInternalServerError)
                log.Println("Failed to read database results:", err)
                return
            }

            // Print each row to the terminal
            // log.Printf("Vendor ID: %d, Vendor Name: %s, Food ID: %d, Food Name: %s, Description: %s, Price: %.2f, Image Path: %s, AddOn ID: %d, AddOn Name: %s, AddOn Price: %.2f\n",
                // vendorId, vendorName, food.ID, food.FoodName, food.Description, food.Price, food.ImagePath, addon.ID, addon.Name, addon.Price)

            // Group add-ons by food item
            if addon.ID != 0 {
                food.Addons = append(food.Addons, addon)
            }

            if vendor, exists := vendors[vendorId]; exists {
                // Check if the food item already exists for the vendor
                foodExists := false
                for i := range vendor.FoodItems {
                    if vendor.FoodItems[i].ID == food.ID {
                        if addon.ID != 0 {
                            vendor.FoodItems[i].Addons = append(vendor.FoodItems[i].Addons, addon)
                        }
                        foodExists = true
                        break
                    }
                }
                if !foodExists {
                    vendor.FoodItems = append(vendor.FoodItems, food)
                }
            } else {
                vendors[vendorId] = &Vendor{
                    ID:        vendorId,
                    Name:      vendorName,
                    FoodItems: []FoodItem{food},
                }
            }
        }

        if err := rows.Err(); err != nil {
            http.Error(w, "Failed to read database results", http.StatusInternalServerError)
            log.Println("Error iterating over rows:", err)
            return
        }

        vendorsList := make([]*Vendor, 0, len(vendors))
        for _, vendor := range vendors {
            vendorsList = append(vendorsList, vendor)
        }

        log.Println("Vendors fetched:", vendorsList)

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(vendorsList); err != nil {
            http.Error(w, "Failed to encode response", http.StatusInternalServerError)
            log.Println("Failed to encode response:", err)
        }
    }
}



func createOrderHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var requestData struct {
            SessionID string    `json:"sessionId"`
            Cart      []CartItem `json:"cart"`
        }

        if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        session, err := store.Get(r, "session-name")
        userId, ok := session.Values["userid"].(int)
        if !ok {
            http.Error(w, "User not logged in", http.StatusUnauthorized)
            return
        }

        if len(requestData.Cart) == 0 {
            http.Error(w, "Cart is empty", http.StatusBadRequest)
            return
        }

        vendorId := requestData.Cart[0].ID // Assuming vendor_id is in cart data
        totalPrice := 0.0
        for _, item := range requestData.Cart {
            totalPrice += item.Price * float64(item.Quantity)
        }

        order := Order{
            UserID:    userId,
            VendorID:  vendorId,
            TotalPrice: totalPrice,
            Status:    "Pending",
        }

        tx, err := db.Begin()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        res, err := tx.Exec("INSERT INTO orders (user_id, vendor_id, total_price, status) VALUES (?, ?, ?, ?)",
            order.UserID, order.VendorID, order.TotalPrice, order.Status)
        if err != nil {
            tx.Rollback()
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        orderId, err := res.LastInsertId()
        if err != nil {
            tx.Rollback()
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        for _, item := range requestData.Cart {
            addonsIds := make([]int, len(item.Addons))
            for i, addon := range item.Addons {
                addonsIds[i] = addon.ID
            }
            addonsIdsJson, err := json.Marshal(addonsIds)
            if err != nil {
                tx.Rollback()
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            _, err = tx.Exec("INSERT INTO order_items (order_id, food_id, quantity, addons_ids, item_price) VALUES (?, ?, ?, ?, ?)",
                orderId, item.ID, item.Quantity, addonsIdsJson, item.Price)
            if err != nil {
                tx.Rollback()
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }

        err = tx.Commit()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{"message": "Order created successfully"})
    }
}
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


func getFoodItemsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("Received request to fetch food items.")
        vendorID := 1

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

// Addon represents an add-on for a food item
type Addon struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
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
            foodItem.Price = 0.0 // Default value if price is empty
        }

        if prepTimeStr != "" {
            foodItem.PrepTime, err = strconv.Atoi(prepTimeStr)
            if err != nil {
                log.Printf("Error converting prep time to int: %v\n", err)
                http.Error(w, "Invalid prep time format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.PrepTime = 0 // Default value if prep time is empty
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

        foodItem.Addons = addons // Correct assignment

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
            foodItem.Price = 0.0 // Default value if price is empty
        }

        if prepTimeStr != "" {
            foodItem.PrepTime, err = strconv.Atoi(prepTimeStr)
            if err != nil {
                log.Printf("Error converting prep time to int: %v\n", err)
                http.Error(w, "Invalid prep time format", http.StatusInternalServerError)
                return
            }
        } else {
            foodItem.PrepTime = 0 // Default value if prep time is empty
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

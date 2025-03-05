package handlers

/*
Sample Input:

POST /execute-select HTTP/1.1
Content-Type: application/json

{
    "table_name": "users"
}

*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQL 驅動程式
)

// Config represents the configuration structure
type Config struct {
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
}

// Global variable to hold the configuration
var config Config

// LoadConfig loads the configuration from a JSON file
func LoadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}
	return nil
}

// TableRequest represents the JSON structure of the incoming request
type TableRequest struct {
	TableName string `json:"table_name"`
}

// ExecuteSelect handles the /execute-select route
func ExecuteSelect(w http.ResponseWriter, r *http.Request) {

    fmt.Println("Run Select Query Called")
	
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body
	var req TableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	fmt.Println("  Request TableName:", req.TableName)

	// Validate the table name
	if req.TableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	// Prevent SQL injection by validating the table name
	if !isValidTableName(req.TableName) {
		http.Error(w, "Invalid table name", http.StatusBadRequest)
		return
	}

	// Connect to the database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Construct the query
	query := fmt.Sprintf("SELECT * FROM %s", req.TableName)

	fmt.Println("  Query:", query)

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusBadRequest)
		fmt.Println("Failed to execute query: ", err)
		return
	}
	defer rows.Close()

	// Parse the query result
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to retrieve columns", http.StatusInternalServerError)
		return
	}

	results := []map[string]interface{}{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}

		// Create a map to hold the column data
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("  Results JSON: ", results)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode results", http.StatusInternalServerError)
		return
	}
}

// isValidTableName validates the table name to prevent SQL injection
func isValidTableName(tableName string) bool {
	// Only allow alphanumeric table names with underscores
	for _, char := range tableName {
		if !(char >= 'a' && char <= 'z') &&
			!(char >= 'A' && char <= 'Z') &&
			!(char >= '0' && char <= '9') &&
			char != '_' {
			return false
		}
	}
	return true
}

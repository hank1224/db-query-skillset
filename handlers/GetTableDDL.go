package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type GetTableDDLRequest struct {
	TableName string `json:"table_name"`
}

func GetTableDDL(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Get Table DDL Called")

	// Ensure only POST requests are allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body
	var req GetTableDDLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	fmt.Println("  Request TableName:", req.TableName)

	// Validate table name
	if req.TableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	// Prevent SQL injection, validate table name
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

	// Query the table's schema and column information to generate the DDL.
	query := `
		SELECT 
			'CREATE TABLE ' || quote_ident(table_schema) || '.' || quote_ident(table_name) || ' (' ||
			array_to_string(
				array_agg(
					quote_ident(column_name) || ' ' || 
					COALESCE(data_type, '') || 
					CASE 
						WHEN character_maximum_length IS NOT NULL THEN '(' || character_maximum_length || ')'
						ELSE '' 
					END ||
					CASE 
						WHEN is_nullable = 'NO' THEN ' NOT NULL'
						ELSE ''
					END
				), 
				', '
			) || 
			');' AS ddl
		FROM 
			information_schema.columns
		WHERE 
			table_schema = 'public' AND 
			table_name = $1
		GROUP BY 
			table_schema, table_name;
	`

	var ddl string
	err = db.QueryRow(query, req.TableName).Scan(&ddl)

	if err != nil {
		http.Error(w, "Failed to get DDL", http.StatusInternalServerError)
		fmt.Println("Failed to get DDL: ", err)
		return
	}

	// Return the DDL as JSON
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("  DDL JSON: ", ddl)

	if err := json.NewEncoder(w).Encode(map[string]string{"ddl": ddl}); err != nil {
		http.Error(w, "Failed to encode DDL", http.StatusInternalServerError)
		return
	}
}
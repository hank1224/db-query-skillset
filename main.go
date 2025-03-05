package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/github/testdatabot/handlers"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	if err := handlers.LoadConfig("config/databaseConfig.json"); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	http.HandleFunc("/_ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	// http.HandleFunc("/random-lorem-ipsum", handlers.Loripsum)

	http.HandleFunc("/run-select-query", handlers.ExecuteSelect)
	http.HandleFunc("/get-table-ddl", handlers.GetTableDDL)

	port := 8080
	fmt.Printf("Server is running on port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}

package main

import (
	"log"
	"net/http"
	"time"

	"backend_machine/config"
	"backend_machine/docs"
	"backend_machine/handlers"
	"backend_machine/repository"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://127.0.0.1:5173": true,
			"http://10.5.0.8:5173":  true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Gagal konek SQL Server: ", err)
	}
	defer db.Close()

	repo := repository.New(db)
	api := handlers.New(repo)

	// Auto update logs_machine setiap 1 menit.
	// Dengan ini logs_machine tetap update walaupun dashboard tidak dibuka.
	startLogsMachineScheduler(repo, 1*time.Minute)

	// Auto logout operator jika mesin offline >= 60 menit.
	// Scheduler berjalan otomatis setiap 5 menit.
	startMachineOperatorAutoLogoutScheduler(repo, 5*time.Minute)

	mux := http.NewServeMux()

	// Swagger
	docs.SwaggerInfo.Host = "localhost:5000"
	docs.SwaggerInfo.BasePath = "/api"

	// Serve swagger docs
	swaggerFiles := http.FileServer(http.Dir("./docs"))
	mux.Handle("/docs/", http.StripPrefix("/docs/", swaggerFiles))

	// Swagger UI handler
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger/", "/swagger/index.html":
			http.ServeFile(w, r, "./docs/swagger.html")
		case "/swagger/doc.json", "/swagger/swagger.json":
			http.ServeFile(w, r, "./docs/swagger.json")
		default:
			http.NotFound(w, r)
		}
	})

	// API Backend
	mux.HandleFunc("/api/health", api.Health)
	mux.HandleFunc("/api/productivity", api.Productivity)
	mux.HandleFunc("/api/process/detail", api.ProcessDetail)
	mux.HandleFunc("/api/machine-settings", api.MachineSettings)
	mux.HandleFunc("/api/employees/search", api.EmployeeSearch)

	// Machine Operator
	mux.HandleFunc("/api/machine-operator/login", api.MachineOperatorLogin)
	mux.HandleFunc("/api/machine-operator/note", api.MachineOperatorNote)
	mux.HandleFunc("/api/machine-operator/active", api.MachineOperatorActive)
	mux.HandleFunc("/api/machine-operator/report", api.MachineOperatorReport)
	mux.HandleFunc("/api/machine-operator/auto-logout/offline", api.MachineOperatorAutoLogoutOffline)

	// Process Style
	mux.HandleFunc("/api/process-style/styles", api.ProcessStyleStyles)
	mux.HandleFunc("/api/process-style/processes", api.ProcessStyleProcesses)
	mux.HandleFunc("/api/process-style/list", api.ProcessStyleList)
	mux.HandleFunc("/api/process-style", api.ProcessStyleCreate)
	mux.HandleFunc("/api/process-style/", api.ProcessStyleByID)

	// WebSocket
	mux.HandleFunc("/ws/productivity", api.ProductivityWS)

	// Static file lama, kalau masih ada folder public
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println("Server jalan: http://localhost:5000")
	log.Println("Endpoint aktif:")
	log.Println("GET    /api/health")
	log.Println("GET    /api/productivity?date=YYYY-MM-DD")
	log.Println("GET    /api/process/detail?uuid=UUID&date=YYYY-MM-DD")
	log.Println("GET    /api/machine-settings")
	log.Println("POST   /api/machine-settings")
	log.Println("PUT    /api/machine-settings")
	log.Println("DELETE /api/machine-settings?uuid=UUID")
	log.Println("GET    /api/employees/search?q=NIK")

	log.Println("POST   /api/machine-operator/login")
	log.Println("POST   /api/machine-operator/note")
	log.Println("GET    /api/machine-operator/active?uuid=UUID")
	log.Println("GET    /api/machine-operator/report?date=YYYY-MM-DD")
	log.Println("POST   /api/machine-operator/auto-logout/offline")

	log.Println("GET    /api/process-style/styles?q=STYLE")
	log.Println("GET    /api/process-style/processes?style=STYLE&q=PROCESS")
	log.Println("GET    /api/process-style/list?q=KEYWORD")
	log.Println("POST   /api/process-style")
	log.Println("PUT    /api/process-style/:id")
	log.Println("DELETE /api/process-style/:id")

	log.Println("WS     /ws/productivity")
	log.Println("")
	log.Println("Swagger UI: http://localhost:5000/swagger/index.html")
	log.Println("Scheduler logs_machine: aktif setiap 1 menit")
	log.Println("Scheduler auto logout offline: aktif setiap 5 menit")

	log.Fatal(http.ListenAndServe(":5000", corsMiddleware(mux)))
}

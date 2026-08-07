package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend_machine/config"
	"backend_machine/docs"
	"backend_machine/handlers"
	"backend_machine/repository"

	"github.com/joho/godotenv"
)

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func parseCSV(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	parts := strings.Split(text, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func buildAllowedOrigins() map[string]bool {
	defaultOrigins := strings.Join([]string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://10.5.0.8:5173",
		"http://10.5.0.8:5175",
		"http://10.5.0.107:5175",
	}, ",")

	origins := parseCSV(getEnv("APP_ALLOWED_ORIGINS", defaultOrigins))
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	return allowed
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := buildAllowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

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
	_ = godotenv.Load()

	appHost := getEnv("APP_HOST", "0.0.0.0")
	appPort := getEnv("APP_PORT", "5000")
	listenAddr := fmt.Sprintf("%s:%s", appHost, appPort)
	publicBaseURL := getEnv("APP_PUBLIC_BASE_URL", fmt.Sprintf("http://localhost:%s", appPort))
	docs.SwaggerInfo.Host = getEnv("APP_SWAGGER_HOST", fmt.Sprintf("localhost:%s", appPort))
	docs.SwaggerInfo.BasePath = "/api"

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Gagal konek SQL Server: ", err)
	}
	defer db.Close()

	repo := repository.New(db)

	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := repo.EnsureMachineSettingSchema(ctx); err != nil {
			log.Println("Peringatan ensure schema machine_setting_manual:", err)
		}
		if err := repo.EnsureLineShiftConfigSchema(ctx); err != nil {
			log.Println("Peringatan ensure schema line_shift_config:", err)
		}
		if err := repo.EnsureMechanicClaimSchema(ctx); err != nil {
			log.Println("Peringatan ensure schema mechanic claim:", err)
		}
		cancel()
	}

	api := handlers.New(repo)

	// Dimatikan sementara: tidak ada API/UI yang membaca dbo.logs_machine.
	// Scheduler ini tiap 1 menit hitung ulang semua mesin + MERGE, membebani SQL.
	//startLogsMachineScheduler(repo, 1*time.Minute)
	startMachineOperatorWorkEndAutoLogoutScheduler(repo, 5*time.Minute)
	//startMachineOperatorAutoLogoutScheduler(repo, 5*time.Minute)

	mux := http.NewServeMux()

	swaggerFiles := http.FileServer(http.Dir("./docs"))
	mux.Handle("/docs/", http.StripPrefix("/docs/", swaggerFiles))
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

	mux.HandleFunc("/api/health", api.Health)
	mux.HandleFunc("/api/productivity", api.Productivity)
	mux.HandleFunc("/api/process/detail", api.ProcessDetail)
	mux.HandleFunc("/api/machine-settings", api.MachineSettings)
	mux.HandleFunc("/api/line-shift-config", api.LineShiftConfig)
	mux.HandleFunc("/api/employees/search", api.EmployeeSearch)

	mux.HandleFunc("/api/machine-operator/login", api.MachineOperatorLogin)
	mux.HandleFunc("/api/machine-operator/note", api.MachineOperatorNote)
	mux.HandleFunc("/api/machine-operator/active", api.MachineOperatorActive)
	mux.HandleFunc("/api/machine-operator/report", api.MachineOperatorReport)
	mux.HandleFunc("/api/machine-operator/auto-logout/offline", api.MachineOperatorAutoLogoutOffline)
	mux.HandleFunc("/api/machine-operator/loss-event/start", api.MachineOperatorLossEventStart)
	mux.HandleFunc("/api/machine-operator/loss-event/finish", api.MachineOperatorLossEventFinish)
	mux.HandleFunc("/api/machine-operator/loss-event/active", api.MachineOperatorLossEventActive)

	mux.HandleFunc("/api/mechanic/identify", api.MechanicIdentify)
	mux.HandleFunc("/api/mechanic/rfid/register", api.MechanicRFIDRegister)
	mux.HandleFunc("/api/mechanic/broken-machines", api.MechanicBrokenList)
	mux.HandleFunc("/api/mechanic/claim", api.MechanicClaim)
	mux.HandleFunc("/api/mechanic/done", api.MechanicDone)

	mux.HandleFunc("/api/process-style/styles", api.ProcessStyleStyles)
	mux.HandleFunc("/api/process-style/processes", api.ProcessStyleProcesses)
	mux.HandleFunc("/api/process-style/list", api.ProcessStyleList)
	mux.HandleFunc("/api/process-style/import", api.ProcessStyleImport)
	mux.HandleFunc("/api/process-style", api.ProcessStyleCreate)
	mux.HandleFunc("/api/process-style/", api.ProcessStyleByID)

	mux.HandleFunc("/ws/productivity", api.ProductivityWS)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println("Server jalan:", publicBaseURL)
	log.Println("Listen address:", listenAddr)
	log.Println("Swagger UI:", publicBaseURL+"/swagger/index.html")
	log.Println("Allowed origins:", strings.Join(parseCSV(getEnv("APP_ALLOWED_ORIGINS", "")), ","))

	log.Fatal(http.ListenAndServe(listenAddr, corsMiddleware(mux)))
}

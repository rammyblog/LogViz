package logviz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rammyblog/logviz/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConfig struct {
	DbUser     string
	DbPassword string
	DbHost     string
	DbName     string
	DbPort     string
}

type Config struct {
	Dsn      string
	DbType   string
	LoginKey string
}

func Init(databaseType string, DbConfig DbConfig) (Config, error) {

	user := DbConfig.DbUser
	password := DbConfig.DbPassword
	host := DbConfig.DbHost
	dbname := DbConfig.DbName
	port := DbConfig.DbPort

	switch databaseType {
	case "mysql":
		return Config{
			DbType: "mysql",
			Dsn:    fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname),
		}, nil

	case "postgres":
		return Config{
			DbType: "postgres",
			Dsn:    fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", user, password, host, port, dbname),
		}, nil

	default:
		return Config{}, errors.New("invalid database type")
	}

}

func connectDb(config Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch config.DbType {
	case "mysql":
		db, err = gorm.Open(mysql.Open(config.Dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Request{})
		return db, nil
	case "postgres":
		db, err = gorm.Open(postgres.Open(config.Dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return db, nil
	default:
		return nil, errors.New("invalid database type")

	}

}

// Logger middleware function
func (config Config) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		headers, err := json.Marshal(r.Header)

		if err != nil {
			fmt.Println(err)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}

		db, err := connectDb(config)
		if err != nil {
			fmt.Println(err)
		}

		// Create a buffer to capture the response body
		buf := &bytes.Buffer{}

		// Create a proxy ResponseWriter to capture the response status code and body
		proxyWriter := &responseWriterProxy{w, http.StatusOK, buf}

		start := time.Now()

		// Serve the request
		next.ServeHTTP(proxyWriter, r)

		// Record the end time
		end := time.Now()

		// Calculate the duration
		duration := end.Sub(start)

		resHeaders, err := json.Marshal(proxyWriter.Header())

		if err != nil {
			fmt.Println(err)
		}

		req := &models.Request{
			ResponseBody:    buf.String(),
			ResponseStatus:  proxyWriter.status,
			ResponseHeaders: string(resHeaders),
			RequestBody:     string(body),
			Path:            r.URL.Path,
			Headers:         string(headers),
			Method:          r.Method,
			Host:            r.Host,
			Ipaddress:       getClientIP(r),
			TimeSpent:       float64(duration),
		}

		err = db.Create(req).Error
		if err != nil {
			fmt.Println(err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

	})

}

// responseWriterProxy is a proxy for http.ResponseWriter that captures the status code and response body
type responseWriterProxy struct {
	http.ResponseWriter
	status int
	buf    *bytes.Buffer
}

// WriteHeader intercepts and stores the status code
func (w *responseWriterProxy) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write intercepts and stores the status code and response body
func (w *responseWriterProxy) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	// Write to the buffer to capture the response body
	n, err := w.buf.Write(data)
	if err != nil {
		return n, err
	}
	// Write to the original ResponseWriter to send the response to the client
	return w.ResponseWriter.Write(data)
}

// getClientIP retrieves the client's IP address from the request headers
func getClientIP(r *http.Request) string {
	// Check if the X-Real-IP header exists
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Check if the X-Forwarded-For header exists
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// The X-Forwarded-For header may contain multiple IP addresses separated by commas,
		// with the client's IP address being the first one
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}
	// If neither header exists, fallback to using the RemoteAddr field
	return strings.Split(r.RemoteAddr, ":")[0]
}

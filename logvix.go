package logviz

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
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

var METHODS = map[string]string{
	"get":    "GET",
	"post":   "POST",
	"delete": "DELETE",
	"option": "OPTION",
	"put":    "PUT",
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	requestQueue = make(chan models.Request)
)

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
			Dsn:    fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, dbname, port),
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
		db.AutoMigrate(&models.Request{})
		return db, nil
	default:
		return nil, errors.New("invalid database type")

	}
}

func (config Config) Serve(port string) error {

	server := http.NewServeMux()

	// Honest to God, I don't know why the usual static serving is not working https://github.com/rammyblog/spendwise/blob/90b9239e97f4500e580fecb062dd3553af725cde/router.go#L28C1-L30C43
	_, filename, _, _ := runtime.Caller(0)
	fileServer := http.FileServer(http.Dir(path.Dir(filename) + "/templates/static/"))
	server.Handle("/static/", http.StripPrefix("/static", fileServer))

	server.HandleFunc("/", config.home)
	server.HandleFunc("/logs", config.Logs)

	server.HandleFunc("/ws", config.serveWs)

	var err error

	go func() {
		err = http.ListenAndServe(port, server)
	}()

	if err != nil {
		return err
	}

	return nil

}

func (config Config) home(w http.ResponseWriter, r *http.Request) {
	var logs []models.Request
	db, err := connectDb(config)
	if err != nil {
		fmt.Println(err)
	}
	// get the last 10 logs
	db.Order("created_at desc").Limit(10).Find(&logs)

	render(w, "index.html", nil)
}

func (config Config) Logs(w http.ResponseWriter, r *http.Request) {
	lastID := r.URL.Query().Get("lastId")
	searchBy := r.URL.Query().Get("searchBy")
	searchTerm := r.URL.Query().Get("searchTerm")
	var path, method, ipAddress string
	var code int

	switch searchBy {
	case "Path":
		path = searchTerm
	case "Code":
		// convert to int
		intCode, err := strconv.Atoi(searchTerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		code = intCode
	case "Method":
		method = METHODS[strings.ToLower(searchTerm)]
	case "Ip Address":
		ipAddress = searchTerm
	}

	var req []models.Request

	request, err := connectDb(config)

	if err != nil {
		fmt.Println(err)
	}
	query := request.Limit(20).Order("id desc")

	if lastID != "0" {
		query.Where("id < ?", lastID)
	}
	query.Where(&models.Request{ResponseStatus: code, Path: path, Method: method, Ipaddress: ipAddress})

	query.Find(&req)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)

	_ = json.NewEncoder(w).Encode(&req)
}

func (config Config) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		logs := <-requestQueue

		reqBodyBytes := new(bytes.Buffer)

		_ = json.NewEncoder(reqBodyBytes).Encode(logs)

		err := conn.WriteJSON(logs)
		if err != nil {
			fmt.Println(err)
			return
		}
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
		proxyWriter := &responseWriterProxy{w, 0, buf}

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
			TimeSpent:       float64(duration.Milliseconds()),
		}

		err = db.Create(req).Error
		if err != nil {
			fmt.Println(err)
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		go func() {
			requestQueue <- *req
		}()
	})

}

// responseWriterProxy is a proxy for http.ResponseWriter that captures the status code and response body
type responseWriterProxy struct {
	http.ResponseWriter
	status int
	buf    *bytes.Buffer
}

// // WriteHeader intercepts and stores the status code
func (w *responseWriterProxy) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write intercepts and stores the status code and response body
func (w *responseWriterProxy) Write(data []byte) (int, error) {
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

	if strings.Contains(r.RemoteAddr, "[::1]") {
		return "127.0.0.1"
	}

	// If neither header exists, fallback to using the RemoteAddr field
	return strings.Split(r.RemoteAddr, ":")[0]
}

//go:embed templates
var templateFS embed.FS

func render(w http.ResponseWriter, t string, data interface{}) {

	// partials := []string{}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("templates/%s", t))

	// for _, x := range partials {
	// 	templateSlice = append(templateSlice, x)
	// }

	tmpl, err := template.ParseFS(templateFS, templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

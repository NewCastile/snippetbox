package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/newcastile/snippetbox/internal/models"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
}

func main() {
	// Command line flags
	addr := flag.String("addr", ":4000", "HTTP network address")

	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	// Logging
	infolog := log.New(os.Stdout, "INFOR\t", log.Ldate|log.Ltime|log.LUTC)
	errorlog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile)

	// Database
	db, err := openDB(*dsn)

	if err != nil {
		errorlog.Fatal(err)
	}

	defer db.Close()

	// Template cache
	templateCache, err := newTemplateCache()

	if err != nil {
		errorlog.Fatal(err)
	}

	// Form decoder
	formDecoder := form.NewDecoder()
	// Session manager
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// App initialization
	app := &application{
		errorLog:       errorlog,
		infoLog:        infolog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
	}

	// Custom TLS Config
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorlog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infolog.Printf("Starting server on port %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorlog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/api"
	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store/mysql"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/namsral/flag"
	"golang.org/x/crypto/bcrypt"
)
var (
	logger       = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
)

func main() {
	portFlag := flag.Int("port", 8001, "port, default is 8001")
	dbHostFlag := flag.String("db_host", "127.0.0.1", "Database host, default is 127.0.0.1")
	dbPortFlag := flag.Int("db_port", 3306, "Database port, default is 3306")
	dbUserFlag := flag.String("db_user", "root", "Database user, default is root")
	dbPasswordFlag := flag.String("db_password", "12345", "Database password, default is 123456")
	dbNameFlag := flag.String("db_name", "go_sample_api_server_structure", "Database name, default is go_sample_api_server_structure")
	flag.Parse()

	port := *portFlag
	dbHost := *dbHostFlag
	dbPort := *dbPortFlag
	dbUser := *dbUserFlag
	dbPassword := *dbPasswordFlag
	dbName := *dbNameFlag

	db, err := mysql.Connect(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		panic(err)
	}

	// For testing purpose add users
	err = addUser(db.User(), "username1", "password1")
	err = addUser(db.User(), "username2", "password2")
	if err != nil {
		panic(fmt.Sprintf("error adding user %v", err))
	}

	apiHandler := api.NewHandler(db, logger)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Mount("/", apiHandler)

	srv := &http.Server{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Addr:         ":" + strconv.Itoa(port),
		Handler:      router,
	}

	fmt.Printf("starting server on port %d\n", port)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("error starting server: %v", err))
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-shutdownSignal:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Shut down gracefully, but wait no longer than 60 seconds before stopping.
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("error shutting down server: %v\n", err)
	}
}

func addUser(us store.UserStore, username, password string) error {
	fmt.Printf("adding username %q with password %q\n", username, password)

	defaultPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = us.Create(context.Background(), username, string(defaultPassword))
	if err != nil && err != store.ErrDuplicate {
		return err
	}

	return nil
}
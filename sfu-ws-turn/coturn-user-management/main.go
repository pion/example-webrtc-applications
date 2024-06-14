package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var (
	dsn         = flag.String("dsn", "/var/lib/turn/turndb", "DSN for the sqlite3 db")
	addr        = flag.String("addr", ":8443", "address to listen to")
	certKeyPath = flag.String("cert-key", "cert-key", "path to the tls certificate private key")
	certPath    = flag.String("cert", "cert", "path to the tls certificate")
	realm       = flag.String("realm", "turn.example.com", "TURN realm")
	url         = flag.String("url", "https://example.com", "URL to allow for CORS")
)

type tempCredentials struct {
	UserName string `json:"Username"`
	Password string `json:"Password`
}

func generateTempCredentialsHash(user string, realm string, password string) (hashStr string) {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", user, realm, password)))

	return hex.EncodeToString(hash[:])
}

func generateTempCredentials(db *sql.DB) (tc tempCredentials, err error) {

	un := fmt.Sprintf("%d", time.Now().UnixNano())
	pw := fmt.Sprintf("%d", rand.Uint64()) // TODO: use crypto/rand instead for prod ready crypto

	tc = tempCredentials{UserName: un, Password: pw}
	hsh := generateTempCredentialsHash(un, *realm, pw)

	err = writeTempCredentials(db, *realm, un, string(hsh))
	if err != nil {
		log.Info(err)
		tc = tempCredentials{}
		return
	}

	return tc, err
}

func writeTempCredentials(db *sql.DB, realm string, username string, hash string) (err error) {

	// write them
	// return them
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO turnusers_lt (realm, name, hmackey) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(realm, username, hash)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func returnTempCredentials(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Infof("Creating credentials for %v", r.RemoteAddr)
	creds, err := generateTempCredentials(db)
	w.Header().Set("Access-Control-Allow-Origin", *url)
	if err != nil {
		log.Info(err)
		w.WriteHeader(500) // Return 500 Internal Server Error.
		return
	}
	json.NewEncoder(w).Encode(creds)
}

func cleanupOldUsers(db *sql.DB) {
	_, err := db.Exec("DELETE FROM turnusers_lt WHERE `name` < strftime('%s', DATE('now', '-1 days'))") // OK
	if err != nil {
		log.Info(err)
	}
}

// loopCleanupOldUsers wipes old users every hour that lasted more than 24hrs
func loopCleanupOldUsers(db *sql.DB) {
	for {
		log.Info("Cleaning up old users")
		cleanupOldUsers(db)
		time.Sleep(60 * 60 * time.Second) // Once an hour
	}
}

func init() {
	log.SetLevel(log.DebugLevel)
	formatter := runtime.Formatter{ChildFormatter: &log.TextFormatter{
		FullTimestamp: true,
	}}
	formatter.Line = true
	log.SetFormatter(&formatter)

}

func main() {
	// Parse the flags passed to program
	flag.Parse()

	log.Print("Connecting to DB")
	database, err := sql.Open("sqlite3", *dsn)

	if err != nil {
		log.Panic(err)
	}

	// credential generation endpoint
	// Wrapping the script to pass the right DB
	// TODO: Something better than this if you're actually using it
	http.HandleFunc("/20987182471824882098", func(w http.ResponseWriter, r *http.Request) { returnTempCredentials(w, r, database) })

	// Cleanup old users in background
	go loopCleanupOldUsers(database)

	log.Print("Starting server")
	// start HTTP server
	log.Fatal(http.ListenAndServeTLS(*addr, *certPath, *certKeyPath, nil))
}

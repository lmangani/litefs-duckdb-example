package main

import (
    "context"
    "database/sql"
    _ "embed"
    "flag"
    "fmt"
    "html/template"
    "log"
    "math/rand"
    "net/http"
    "os"
    "time"

    "github.com/brianvoe/gofakeit/v6"
    _ "github.com/marcboeker/go-duckdb"
)

// Command line flags.
var (
    dsn  = flag.String("dsn", "memory", "datasource name")
    addr = flag.String("addr", ":8080", "bind address")
)

var db *sql.DB

//go:embed schema.sql
var schemaSQL string

func main() {
    log.SetFlags(0)
    rand.Seed(time.Now().UnixNano())

    if err := run(context.Background()); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run(ctx context.Context) (err error) {
    flag.Parse()

    if *dsn == "" {
        return fmt.Errorf("dsn required")
    } else if *addr == "" {
        return fmt.Errorf("bind address required")
    }

    // Connect to DuckDB database with memory storage.
    // db, err = sql.Open("duckdb", *dsn)
    db, err = sql.Open("duckdb", "")
    if err != nil {
        return fmt.Errorf("open db: %w", err)
    }
    defer db.Close()

    // Ensure the database is reachable.
    if err := db.Ping(); err != nil {
        return fmt.Errorf("ping db: %w", err)
    }

    log.Printf("database opened with memory storage")

    // Create a connection from the database pool.
    conn, err := db.Conn(ctx)
    if err != nil {
        return fmt.Errorf("get db connection: %w", err)
    }
    defer conn.Close()

    // Log the current access mode.
    setting := conn.QueryRowContext(ctx, "SELECT current_setting('access_mode')")
    var accessMode string
    if err := setting.Scan(&accessMode); err != nil {
        return fmt.Errorf("get access mode: %w", err)
    }
    log.Printf("DB opened with access mode %s", accessMode)

    // Install and load the SQLite extension.
    if _, err := conn.ExecContext(ctx, "INSTALL sqlite; LOAD sqlite;"); err != nil {
        return fmt.Errorf("cannot install/load sqlite extension: %w", err)
    } else {
        log.Printf("INSTALL OK")
    }

    if _, err := conn.ExecContext(ctx, "ATTACH '/litefs/db.sqlite' (TYPE SQLITE); USE db;"); err != nil {
        return fmt.Errorf("cannot install/load sqlite extension: %w", err)
    } else {
        log.Printf("ATTACH OK")
    }

    if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
        return fmt.Errorf("cannot migrate schema: %w", err)
    } else {
        log.Printf("CREATE SQL OK")
    }

    // Start HTTP server.
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/generate", handleGenerate)

    log.Printf("http server listening on %s", *addr)
    return http.ListenAndServe(*addr, nil)
}

//go:embed index.tmpl
var indexTmplContent string
var indexTmpl = template.Must(template.New("index").Parse(indexTmplContent))

func handleIndex(w http.ResponseWriter, r *http.Request) {
    // If a different region is specified, redirect to that region.
    if region := r.URL.Query().Get("region"); region != "" && region != os.Getenv("FLY_REGION") {
        log.Printf("redirecting from %q to %q", os.Getenv("FLY_REGION"), region)
        w.Header().Set("fly-replay", "region="+region)
        return
    }

    // Query for the most recently added people.
    rows, err := db.Query(`
        SELECT id, name, phone, company
        FROM db.persons
        ORDER BY id DESC
        LIMIT 10
    `)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Collect rows into a slice.
    var persons []*Person
    for rows.Next() {
        var person Person
        if err := rows.Scan(&person.ID, &person.Name, &person.Phone, &person.Company); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        persons = append(persons, &person)
    }
    if err := rows.Close(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Render the list to either text or HTML.
    tmplData := TemplateData{
        Region:  os.Getenv("FLY_REGION"),
        Persons: persons,
    }

    switch r.Header.Get("accept") {
    case "text/plain":
        fmt.Fprintf(w, "REGION: %s\n\n", tmplData.Region)
        for _, person := range tmplData.Persons {
            fmt.Fprintf(w, "- %s @ %s (%s)\n", person.Name, person.Company, person.Phone)
        }

    default:
        if err := indexTmpl.ExecuteTemplate(w, "index", tmplData); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
    // Only allow POST methods.
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var nextID int
    err := db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(id), 0) + 1 AS next_id FROM db.persons").Scan(&nextID)
    if err != nil {
        fmt.Fprintf(w, "ERROR: %s\n\n", err)
        http.Error(w, "Failed to calculate next ID", http.StatusInternalServerError)
        return
    }

    // If this is the primary, attempt to write a record to the database.
    person := Person{
        Name:    gofakeit.Name(),
        Phone:   gofakeit.Phone(),
        Company: gofakeit.Company(),
    }

    if _, err := db.ExecContext(r.Context(), `INSERT INTO db.persons (id, name, phone, company) VALUES (?, ?, ?, ?)`, nextID, person.Name, person.Phone, person.Company); err != nil {
        fmt.Fprintf(w, "ERROR: %s\n\n", err)
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Redirect back to the index page to view the new result.
    http.Redirect(w, r, r.Referer(), http.StatusFound)
}

type TemplateData struct {
    Region  string
    Persons []*Person
}

type Person struct {
    ID      int
    Name    string
    Phone   string
    Company string
}

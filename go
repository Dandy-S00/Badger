// main.go — compact version
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "sync"
    "time"

    "github.com/gorilla/mux"
    "github.com/rs/cors"
)

type RouteMetadata struct {
    Method       string   `json:"method"`
    Path         string   `json:"path"`
    HandlerName  string   `json:"handlerName"`
    Description  string   `json:"description,omitempty"`
    AuthRequired bool     `json:"authRequired"`
    RateLimit    int      `json:"rateLimit,omitempty"`
    Tags         []string `json:"tags,omitempty"`
    CreatedAt    time.Time `json:"createdAt"`
}

type APIRegistry struct {
    mu     sync.RWMutex
    routes []RouteMetadata
}

func (r *APIRegistry) Register(method, path, handler, description string, auth bool, rateLimit int, tags []string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.routes = append(r.routes, RouteMetadata{
        Method:       method, Path: path, HandlerName: handler,
        Description: description, AuthRequired: auth, RateLimit: rateLimit, Tags: tags,
        CreatedAt: time.Now().UTC(),
    })
}

func (r *APIRegistry) List() []RouteMetadata {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]RouteMetadata, len(r.routes))
    copy(out, r.routes)
    return out
}

var registry = &APIRegistry{}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !strings.HasPrefix(token, "Bearer ") {
            http.Error(w, `{"error": "missing or invalid Authorization header"}`, http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "time": time.Now().UTC().Format(time.RFC3339)})
}

func apisHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "name": "Badger API Gateway", "version": "1.0.0",
        "count": len(registry.List()), "routes": registry.List(),
    })
}

func openAPIHandler(w http.ResponseWriter, r *http.Request) {
    paths := make(map[string]interface{})
    for _, route := range registry.List() {
        path := route.Path
        if _, ok := paths[path]; !ok { paths[path] = make(map[string]interface{}) }
        for _, method := range strings.Split(strings.ToLower(route.Method), ",") {
            op := map[string]interface{}{"summary": route.HandlerName, "description": route.Description}
            if route.AuthRequired { op["security"] = []map[string][]string{{"bearerAuth": {}}} }
            op["responses"] = map[string]string{"200": "OK"}
            paths[path].(map[string]interface{})[method] = op
        }
    }
    spec := map[string]interface{}{"openapi": "3.0.3", "info": map[string]string{"title": "Badger", "version": "1.0.0"}, "paths": paths}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(spec)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    switch r.Method {
    case "GET": json.NewEncoder(w).Encode(map[string]interface{}{"users": []map[string]int{{"id": 1}, {"id": 2}}})
    case "POST": w.WriteHeader(http.StatusCreated); w.Write([]byte(`{"status":"created"}`))
    default: http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
    }
}

func main() {
    registry.Register("GET", "/health", "HealthCheck", "Service health", false, 0, []string{"system"})
    registry.Register("GET", "/apis", "ListAPIs", "Discover endpoints", false, 0, []string{"discovery"})
    registry.Register("GET", "/openapi.json", "OpenAPI", "Spec generation", false, 0, []string{"discovery"})
    registry.Register("GET,POST", "/api/v1/users", "Users", "User CRUD", true, 100, []string{"users"})

    r := mux.NewRouter()
    r.HandleFunc("/health", healthHandler).Methods("GET")
    r.HandleFunc("/apis", apisHandler).Methods("GET")
    r.HandleFunc("/openapi.json", openAPIHandler).Methods("GET")
    protected := r.NewRoute().Subrouter()
    protected.Use(authMiddleware)
    protected.HandleFunc("/api/v1/users", usersHandler).Methods("GET", "POST")
    srv := &http.Server{Addr: ":8080", Handler: cors.Default().Handler(r)}
    go func() { log.Println("Badger starting on :8080"); srv.ListenAndServe() }()
    <-make(chan os.Signal, 1)
    ctx, c := context.WithTimeout(context.Background(), 10*time.Second); defer c()
    srv.Shutdown(ctx)
}

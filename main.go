cat > main.go <<'EOF'
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

var client *firestore.Client

type Customer struct {
	Email string `firestore:"email"`
	ID    string `firestore:"id"`
	Name  string `firestore:"name"`
	Phone string `firestore:"phone"`
}

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("PROJECT_ID")
	}
	if projectID == "" {
		projectID = "YOUR_PROJECT_ID"
	}

	var err error
	client, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Error initializing Cloud Firestore client: %v", err)
	}
	defer client.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/", rootHandler).Methods("GET", "HEAD")
	r.HandleFunc("/v1/customer/{id}", customerHandler).Methods("GET", "OPTIONS")

	corsMiddleware := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Origin", "Content-Type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "OPTIONS"}),
	)

	log.Println("Pets REST API listening on port", port)
	handler := corsMiddleware(r)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error launching Pets REST API server: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "running"})
}

func customerHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	ctx := context.Background()

	cust, err := getCustomer(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "fail", "data": fmt.Sprintf("Error fetching customer: %v", err)})
		return
	}
	if cust == nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"status": "fail", "data": map[string]string{"title": fmt.Sprintf("Customer \"%s\" not found", id)}})
		return
	}

	amounts, err := getAmounts(ctx, cust)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "fail", "data": fmt.Sprintf("Unable to fetch amounts: %v", err)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": amounts})
}

func getCustomer(ctx context.Context, id string) (*Customer, error) {
	q := client.Collection("customers").Where("id", "==", id).Limit(1)
	iter := q.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var c Customer
	if err := doc.DataTo(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func getAmounts(ctx context.Context, c *Customer) (map[string]int64, error) {
	if c == nil {
		return map[string]int64{}, fmt.Errorf("Customer is nil")
	}
	result := map[string]int64{
		"proposed": 0,
		"approved": 0,
		"rejected": 0,
	}

	collPath := fmt.Sprintf("customers/%s/treatments", c.Email)
	iter := client.Collection(collPath).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result, err
		}
		data := doc.Data()

		statusRaw, ok := data["status"]
		if !ok {
			continue
		}
		status, ok := statusRaw.(string)
		if !ok {
			continue
		}

		var costInt64 int64
		if costRaw, ok := data["cost"]; ok {
			switch v := costRaw.(type) {
			case int64:
				costInt64 = v
			case int:
				costInt64 = int64(v)
			case float64:
				costInt64 = int64(v)
			case float32:
				costInt64 = int64(v)
			default:
				continue
			}
		} else {
			continue
		}

		if _, exists := result[status]; exists {
			result[status] += costInt64
		}
	}

	return result, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
EOF

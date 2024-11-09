package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/naanf-akma/v1/register-anchorkey", RegisterAKMAKey)
	http.HandleFunc("/naanf-akma/v1/retrieve-applicationkey", GetAKMAAPPKeyMaterial)
	http.HandleFunc("/naanf-akma/v1/remove-context", RemoveContext)

	log.Println("Starting server on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func RegisterAKMAKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var akmaKeyInfo AkmaKeyInfo
	if err := json.NewDecoder(r.Body).Decode(&akmaKeyInfo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process the AKMA key info here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(akmaKeyInfo)
}

func GetAKMAAPPKeyMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var keyRequest AkmaAfKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&keyRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch and return AKMA Application Key Material
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AkmaAfKeyData{})
}

func RemoveContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var ctxRemove CtxRemove
	if err := json.NewDecoder(r.Body).Decode(&ctxRemove); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Handle the context removal
	w.WriteHeader(http.StatusNoContent)
}

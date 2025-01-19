package main

import (
	"net/http"
	"log"
	"encoding/json"
	"testing"
)

func TestSignUp(t *testing.T) {
	req, _ := http.NewRequest("POST", "/sign-up", nil)
	w := httptest.NewRecorder()
	signUp(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLogin(t *testing.T) {
	req, _ := http.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()
	login(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
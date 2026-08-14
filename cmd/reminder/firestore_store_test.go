package reminder

import (
	"context"
	"strings"
	"testing"
)

func Test_isFirestoreConfigured_emulatorHost(t *testing.T) {
	t.Setenv(envFirestoreServiceAccountJSONFile, "")
	t.Setenv(envFirestoreEmulatorHost, "localhost:8080")
	if !isFirestoreConfigured() {
		t.Fatalf("expected firestore configured when emulator host is set")
	}
}

func Test_isFirestoreConfigured_credentialsFile(t *testing.T) {
	t.Setenv(envFirestoreEmulatorHost, "")
	t.Setenv(envFirestoreServiceAccountJSONFile, "/tmp/creds.json")
	if !isFirestoreConfigured() {
		t.Fatalf("expected firestore configured when credentials file is set")
	}
}

func Test_loadFirestoreEnvConfig(t *testing.T) {
	t.Setenv(envFirestoreProjectID, "proj")
	t.Setenv(envFirestoreServiceAccountJSONFile, "/tmp/key.json")
	t.Setenv(envFirestoreEmulatorHost, "127.0.0.1:8080")

	cfg := loadFirestoreEnvConfig()
	if cfg.projectID != "proj" {
		t.Fatalf("unexpected projectID: %q", cfg.projectID)
	}
	if cfg.credentialsFile != "/tmp/key.json" {
		t.Fatalf("unexpected credentialsFile: %q", cfg.credentialsFile)
	}
	if cfg.emulatorHost != "127.0.0.1:8080" {
		t.Fatalf("unexpected emulatorHost: %q", cfg.emulatorHost)
	}
}

func Test_newFirestoreClient_requiresProjectID(t *testing.T) {
	cfg := firestoreEnvConfig{
		projectID:       "",
		credentialsFile: "/tmp/key.json",
		emulatorHost:    "",
	}

	_, err := newFirestoreClient(context.Background(), cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), envFirestoreProjectID) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_newFirestoreClient_requiresCredentialFile(t *testing.T) {
	cfg := firestoreEnvConfig{
		projectID:       "proj",
		credentialsFile: "",
		emulatorHost:    "",
	}

	_, err := newFirestoreClient(context.Background(), cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), envFirestoreServiceAccountJSONFile) {
		t.Fatalf("unexpected error: %v", err)
	}
}

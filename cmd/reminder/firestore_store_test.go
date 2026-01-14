package reminder

import (
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

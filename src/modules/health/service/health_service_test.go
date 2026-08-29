package service

import "testing"

func TestHealthService_Check_Structure(t *testing.T) {
	svc := NewHealthService("test-app")
	result := svc.Check()

	if result["app"] != "test-app" {
		t.Errorf("app = %v, want test-app", result["app"])
	}
	if result["status"] != "degraded" {
		t.Errorf("status = %v, want degraded", result["status"])
	}
	if _, ok := result["database"]; !ok {
		t.Error("missing database key in response")
	}
	if _, ok := result["redis"]; !ok {
		t.Error("missing redis key in response")
	}
}

func TestNewHealthService_StoresAppName(t *testing.T) {
	svc := NewHealthService("my-service")
	if svc.appName != "my-service" {
		t.Errorf("appName = %q, want my-service", svc.appName)
	}
}

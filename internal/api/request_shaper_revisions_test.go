package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func TestRequestShaper_ShapeRevisionsGet(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file123", "key123", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Revisions.Get("file123", "rev456")
	shaped := shaper.ShapeRevisionsGet(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file123/key123" {
		t.Errorf("Expected resource key header 'file123/key123', got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsGet_NoResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	// Don't add any resource keys

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Revisions.Get("file123", "rev456")
	shaped := shaper.ShapeRevisionsGet(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "" {
		t.Errorf("Expected empty header, got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsUpdate(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file123", "key123", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	rev := &drive.Revision{KeepForever: true}
	call := service.Revisions.Update("file123", "rev456", rev)
	shaped := shaper.ShapeRevisionsUpdate(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file123/key123" {
		t.Errorf("Expected resource key header 'file123/key123', got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsUpdate_NoResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	// Don't add any resource keys

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	rev := &drive.Revision{KeepForever: true}
	call := service.Revisions.Update("file123", "rev456", rev)
	shaped := shaper.ShapeRevisionsUpdate(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "" {
		t.Errorf("Expected empty header, got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsGet_MultipleFiles(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file1", "key1", "test")
	client.ResourceKeys().AddKey("file2", "key2", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file1", "file2"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Revisions.Get("file1", "rev1")
	shaped := shaper.ShapeRevisionsGet(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	// Should include both file resource keys
	if header != "file1/key1,file2/key2" {
		t.Errorf("Expected resource key header 'file1/key1,file2/key2', got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsUpdate_MultipleFiles(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file1", "key1", "test")
	client.ResourceKeys().AddKey("file2", "key2", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file1", "file2"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	rev := &drive.Revision{KeepForever: true}
	call := service.Revisions.Update("file1", "rev1", rev)
	shaped := shaper.ShapeRevisionsUpdate(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	// Should include both file resource keys
	if header != "file1/key1,file2/key2" {
		t.Errorf("Expected resource key header 'file1/key1,file2/key2', got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsGet_HeaderMethod(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Revisions.Get("file123", "rev456")
	shaped := shaper.ShapeRevisionsGet(call, reqCtx)

	// Verify Header() method works
	header := shaped.Header()
	if header == nil {
		t.Error("Header() returned nil")
	}
}

func TestRequestShaper_ShapeRevisionsUpdate_HeaderMethod(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	rev := &drive.Revision{KeepForever: true}
	call := service.Revisions.Update("file123", "rev456", rev)
	shaped := shaper.ShapeRevisionsUpdate(call, reqCtx)

	// Verify Header() method works
	header := shaped.Header()
	if header == nil {
		t.Error("Header() returned nil")
	}
}

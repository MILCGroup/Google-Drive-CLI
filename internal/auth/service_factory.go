package auth

import (
	"context"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/script/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/slides/v1"
	"google.golang.org/api/tasks/v1"
)

type ServiceType string

const (
	ServiceDrive         ServiceType = "drive"
	ServiceSheets        ServiceType = "sheets"
	ServiceDocs          ServiceType = "docs"
	ServiceSlides        ServiceType = "slides"
	ServiceAdminDir      ServiceType = "admin_directory"
	ServiceChat          ServiceType = "chat"
	ServiceGmail         ServiceType = "gmail"
	ServiceCalendar      ServiceType = "calendar"
	ServicePeople        ServiceType = "people"
	ServiceTasks         ServiceType = "tasks"
	ServiceForms         ServiceType = "forms"
	ServiceAppScript     ServiceType = "appscript"
	ServiceCloudIdentity ServiceType = "cloudidentity"
)

type ServiceFactory struct {
	manager *Manager
}

func NewServiceFactory(manager *Manager) *ServiceFactory {
	return &ServiceFactory{manager: manager}
}

func (f *ServiceFactory) CreateService(ctx context.Context, creds *types.Credentials, svcType ServiceType) (interface{}, error) {
	switch svcType {
	case ServiceDrive:
		return f.CreateDriveService(ctx, creds)
	case ServiceSheets:
		return f.CreateSheetsService(ctx, creds)
	case ServiceDocs:
		return f.CreateDocsService(ctx, creds)
	case ServiceSlides:
		return f.CreateSlidesService(ctx, creds)
	case ServiceAdminDir:
		return f.CreateAdminService(ctx, creds)
	case ServiceChat:
		return f.CreateChatService(ctx, creds)
	case ServiceGmail:
		return f.CreateGmailService(ctx, creds)
	case ServiceCalendar:
		return f.CreateCalendarService(ctx, creds)
	case ServicePeople:
		return f.CreatePeopleService(ctx, creds)
	case ServiceTasks:
		return f.CreateTasksService(ctx, creds)
	case ServiceForms:
		return f.CreateFormsService(ctx, creds)
	case ServiceAppScript:
		return f.CreateAppScriptService(ctx, creds)
	case ServiceCloudIdentity:
		return f.CreateCloudIdentityService(ctx, creds)
	default:
		return nil, fmt.Errorf("unknown service type: %s", svcType)
	}
}

func (f *ServiceFactory) CreateDriveService(ctx context.Context, creds *types.Credentials) (*drive.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return drive.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateSheetsService(ctx context.Context, creds *types.Credentials) (*sheets.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateDocsService(ctx context.Context, creds *types.Credentials) (*docs.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return docs.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateSlidesService(ctx context.Context, creds *types.Credentials) (*slides.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return slides.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateAdminService(ctx context.Context, creds *types.Credentials) (*admin.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return admin.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateChatService(ctx context.Context, creds *types.Credentials) (*chat.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return chat.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateGmailService(ctx context.Context, creds *types.Credentials) (*gmail.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateCalendarService(ctx context.Context, creds *types.Credentials) (*calendar.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreatePeopleService(ctx context.Context, creds *types.Credentials) (*people.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return people.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateTasksService(ctx context.Context, creds *types.Credentials) (*tasks.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return tasks.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateFormsService(ctx context.Context, creds *types.Credentials) (*forms.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return forms.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateAppScriptService(ctx context.Context, creds *types.Credentials) (*script.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return script.NewService(ctx, option.WithHTTPClient(client))
}

func (f *ServiceFactory) CreateCloudIdentityService(ctx context.Context, creds *types.Credentials) (*cloudidentity.Service, error) {
	client := f.manager.GetHTTPClient(ctx, creds)
	return cloudidentity.NewService(ctx, option.WithHTTPClient(client))
}

package meet

import (
	"context"
	"fmt"

	meetapi "cloud.google.com/go/apps/meet/apiv2"
	"cloud.google.com/go/apps/meet/apiv2/meetpb"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Manager handles Google Meet API operations via gRPC
type Manager struct {
	spacesClient            *meetapi.SpacesClient
	conferenceRecordsClient *meetapi.ConferenceRecordsClient
}

// NewManager creates a new Meet manager with gRPC clients
func NewManager(ctx context.Context, opts ...option.ClientOption) (*Manager, error) {
	spacesClient, err := meetapi.NewSpacesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create spaces client: %w", err)
	}

	conferenceRecordsClient, err := meetapi.NewConferenceRecordsClient(ctx, opts...)
	if err != nil {
		if spacesClient != nil {
			spacesClient.Close()
		}
		return nil, fmt.Errorf("failed to create conference records client: %w", err)
	}

	return &Manager{
		spacesClient:            spacesClient,
		conferenceRecordsClient: conferenceRecordsClient,
	}, nil
}

// Close closes the gRPC connections
func (m *Manager) Close() error {
	var spacesErr, recordsErr error
	if m.spacesClient != nil {
		spacesErr = m.spacesClient.Close()
	}
	if m.conferenceRecordsClient != nil {
		recordsErr = m.conferenceRecordsClient.Close()
	}
	if spacesErr != nil {
		return spacesErr
	}
	return recordsErr
}

// GetSpace gets a specific Meet space
func (m *Manager) GetSpace(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.MeetSpace, error) {
	req := &meetpb.GetSpaceRequest{
		Name: name,
	}

	resp, err := m.spacesClient.GetSpace(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get space: %w", err)
	}

	space := convertSpace(resp)
	return &space, nil
}

// CreateSpace creates a new Meet space
func (m *Manager) CreateSpace(ctx context.Context, reqCtx *types.RequestContext, accessType string) (*types.MeetCreateSpaceResponse, error) {
	space := &meetpb.Space{}
	if accessType != "" {
		space.Config = &meetpb.SpaceConfig{
			AccessType: meetpb.SpaceConfig_AccessType(meetpb.SpaceConfig_AccessType_value[accessType]),
		}
	}

	req := &meetpb.CreateSpaceRequest{
		Space: space,
	}

	resp, err := m.spacesClient.CreateSpace(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	return &types.MeetCreateSpaceResponse{
		Name:        resp.Name,
		MeetingUri:  resp.MeetingUri,
		MeetingCode: resp.MeetingCode,
	}, nil
}

// UpdateSpace updates a Meet space
func (m *Manager) UpdateSpace(ctx context.Context, reqCtx *types.RequestContext, name, accessType string) (*types.MeetSpace, error) {
	space := &meetpb.Space{
		Name: name,
	}
	if accessType != "" {
		space.Config = &meetpb.SpaceConfig{
			AccessType: meetpb.SpaceConfig_AccessType(meetpb.SpaceConfig_AccessType_value[accessType]),
		}
	}

	req := &meetpb.UpdateSpaceRequest{
		Space: space,
	}

	resp, err := m.spacesClient.UpdateSpace(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update space: %w", err)
	}

	spaceResult := convertSpace(resp)
	return &spaceResult, nil
}

// EndActiveConference ends the active conference in a space
func (m *Manager) EndActiveConference(ctx context.Context, reqCtx *types.RequestContext, name string) error {
	req := &meetpb.EndActiveConferenceRequest{
		Name: name,
	}

	if err := m.spacesClient.EndActiveConference(ctx, req); err != nil {
		return fmt.Errorf("failed to end active conference: %w", err)
	}

	return nil
}

// ListConferenceRecords lists conference records
func (m *Manager) ListConferenceRecords(ctx context.Context, reqCtx *types.RequestContext, pageSize int32, pageToken, filter string) (*types.MeetConferenceRecordsListResponse, error) {
	req := &meetpb.ListConferenceRecordsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
	}

	iter := m.conferenceRecordsClient.ListConferenceRecords(ctx, req)
	var records []types.MeetConferenceRecord
	for {
		record, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list conference records: %w", err)
		}
		records = append(records, convertConferenceRecord(record))
	}

	return &types.MeetConferenceRecordsListResponse{
		ConferenceRecords: records,
	}, nil
}

// GetConferenceRecord gets a specific conference record
func (m *Manager) GetConferenceRecord(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.MeetConferenceRecord, error) {
	req := &meetpb.GetConferenceRecordRequest{
		Name: name,
	}

	resp, err := m.conferenceRecordsClient.GetConferenceRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get conference record: %w", err)
	}

	record := convertConferenceRecord(resp)
	return &record, nil
}

// ListParticipants lists participants in a conference
func (m *Manager) ListParticipants(ctx context.Context, reqCtx *types.RequestContext, parent string, pageSize int32, pageToken string) (*types.MeetParticipantsListResponse, error) {
	req := &meetpb.ListParticipantsRequest{
		Parent:    parent,
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	iter := m.conferenceRecordsClient.ListParticipants(ctx, req)
	var participants []types.MeetParticipant
	for {
		participant, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list participants: %w", err)
		}
		participants = append(participants, convertParticipant(participant))
	}

	return &types.MeetParticipantsListResponse{
		Participants: participants,
	}, nil
}

// GetParticipant gets a specific participant
func (m *Manager) GetParticipant(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.MeetParticipant, error) {
	req := &meetpb.GetParticipantRequest{
		Name: name,
	}

	resp, err := m.conferenceRecordsClient.GetParticipant(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	participant := convertParticipant(resp)
	return &participant, nil
}

func convertSpace(space *meetpb.Space) types.MeetSpace {
	s := types.MeetSpace{
		Name:        space.Name,
		MeetingUri:  space.MeetingUri,
		MeetingCode: space.MeetingCode,
	}
	if space.Config != nil {
		s.Config = &types.MeetSpaceConfig{
			AccessType: space.Config.AccessType.String(),
		}
	}
	if space.ActiveConference != nil {
		s.ActiveConference = space.ActiveConference.ConferenceRecord
	}
	return s
}

func convertConferenceRecord(record *meetpb.ConferenceRecord) types.MeetConferenceRecord {
	r := types.MeetConferenceRecord{
		Name:  record.Name,
		Space: record.Space,
	}
	if record.StartTime != nil {
		r.StartTime = record.StartTime.AsTime().Format("2006-01-02 15:04:05")
	}
	if record.EndTime != nil {
		r.EndTime = record.EndTime.AsTime().Format("2006-01-02 15:04:05")
	}
	if record.ExpireTime != nil {
		r.ExpireTime = record.ExpireTime.AsTime().Format("2006-01-02 15:04:05")
	}
	return r
}

func convertParticipant(participant *meetpb.Participant) types.MeetParticipant {
	p := types.MeetParticipant{
		Name: participant.Name,
	}

	switch u := participant.User.(type) {
	case *meetpb.Participant_SignedinUser:
		p.DisplayName = u.SignedinUser.DisplayName
		p.Email = u.SignedinUser.User
	case *meetpb.Participant_AnonymousUser:
		p.DisplayName = u.AnonymousUser.DisplayName
	case *meetpb.Participant_PhoneUser:
		p.DisplayName = u.PhoneUser.DisplayName
		p.PhoneNumber = "Phone"
	}

	if participant.EarliestStartTime != nil {
		p.EarliestStartTime = participant.EarliestStartTime.AsTime().Format("2006-01-02 15:04:05")
	}
	if participant.LatestEndTime != nil {
		p.LatestEndTime = participant.LatestEndTime.AsTime().Format("2006-01-02 15:04:05")
	}
	return p
}

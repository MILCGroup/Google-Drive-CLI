package types

// MeetSpace represents a Google Meet space
type MeetSpace struct {
	Name             string           `json:"name"`
	MeetingUri       string           `json:"meetingUri,omitempty"`
	MeetingCode      string           `json:"meetingCode,omitempty"`
	Config           *MeetSpaceConfig `json:"config,omitempty"`
	ActiveConference string           `json:"activeConference,omitempty"`
}

func (s *MeetSpace) Headers() []string {
	return []string{"Space Name", "Meeting URI", "Meeting Code"}
}

func (s *MeetSpace) Rows() [][]string {
	return [][]string{{
		truncateID(s.Name, 40),
		s.MeetingUri,
		s.MeetingCode,
	}}
}

func (s *MeetSpace) EmptyMessage() string {
	return "No space information available"
}

// MeetSpaceConfig represents space configuration
type MeetSpaceConfig struct {
	AccessType string `json:"accessType,omitempty"`
}

// MeetConferenceRecord represents a conference recording
type MeetConferenceRecord struct {
	Name       string `json:"name"`
	StartTime  string `json:"startTime,omitempty"`
	EndTime    string `json:"endTime,omitempty"`
	ExpireTime string `json:"expireTime,omitempty"`
	Space      string `json:"space,omitempty"`
}

func (r *MeetConferenceRecord) Headers() []string {
	return []string{"Record Name", "Start Time", "End Time", "Space"}
}

func (r *MeetConferenceRecord) Rows() [][]string {
	return [][]string{{
		truncateID(r.Name, 40),
		r.StartTime,
		r.EndTime,
		truncateID(r.Space, 30),
	}}
}

func (r *MeetConferenceRecord) EmptyMessage() string {
	return "No conference record found"
}

// MeetParticipant represents a meeting participant
type MeetParticipant struct {
	Name              string `json:"name"`
	DisplayName       string `json:"displayName,omitempty"`
	Email             string `json:"email,omitempty"`
	PhoneNumber       string `json:"phoneNumber,omitempty"`
	EarliestStartTime string `json:"earliestStartTime,omitempty"`
	LatestEndTime     string `json:"latestEndTime,omitempty"`
}

func (p *MeetParticipant) Headers() []string {
	return []string{"Participant Name", "Display Name", "Email"}
}

func (p *MeetParticipant) Rows() [][]string {
	return [][]string{{
		truncateID(p.Name, 40),
		p.DisplayName,
		p.Email,
	}}
}

func (p *MeetParticipant) EmptyMessage() string {
	return "No participant information available"
}

// MeetSpacesListResponse represents a list of spaces response
type MeetSpacesListResponse struct {
	Spaces        []MeetSpace `json:"spaces"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

func (r *MeetSpacesListResponse) Headers() []string {
	return []string{"Space Name", "Meeting URI", "Meeting Code"}
}

func (r *MeetSpacesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Spaces))
	for i, space := range r.Spaces {
		rows[i] = []string{
			truncateID(space.Name, 40),
			space.MeetingUri,
			space.MeetingCode,
		}
	}
	return rows
}

func (r *MeetSpacesListResponse) EmptyMessage() string {
	return "No spaces found"
}

// MeetConferenceRecordsListResponse represents a list of conference records
type MeetConferenceRecordsListResponse struct {
	ConferenceRecords []MeetConferenceRecord `json:"conferenceRecords"`
	NextPageToken     string                 `json:"nextPageToken,omitempty"`
}

func (r *MeetConferenceRecordsListResponse) Headers() []string {
	return []string{"Record Name", "Start Time", "End Time", "Space"}
}

func (r *MeetConferenceRecordsListResponse) Rows() [][]string {
	rows := make([][]string, len(r.ConferenceRecords))
	for i, record := range r.ConferenceRecords {
		rows[i] = []string{
			truncateID(record.Name, 40),
			record.StartTime,
			record.EndTime,
			truncateID(record.Space, 30),
		}
	}
	return rows
}

func (r *MeetConferenceRecordsListResponse) EmptyMessage() string {
	return "No conference records found"
}

// MeetParticipantsListResponse represents a list of participants
type MeetParticipantsListResponse struct {
	Participants  []MeetParticipant `json:"participants"`
	NextPageToken string            `json:"nextPageToken,omitempty"`
}

func (r *MeetParticipantsListResponse) Headers() []string {
	return []string{"Participant Name", "Display Name", "Email"}
}

func (r *MeetParticipantsListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Participants))
	for i, participant := range r.Participants {
		rows[i] = []string{
			truncateID(participant.Name, 40),
			participant.DisplayName,
			participant.Email,
		}
	}
	return rows
}

func (r *MeetParticipantsListResponse) EmptyMessage() string {
	return "No participants found"
}

// MeetCreateSpaceResponse represents a create space response
type MeetCreateSpaceResponse struct {
	Name        string `json:"name"`
	MeetingUri  string `json:"meetingUri"`
	MeetingCode string `json:"meetingCode"`
}

func (r *MeetCreateSpaceResponse) Headers() []string {
	return []string{"Space Name", "Meeting URI", "Meeting Code"}
}

func (r *MeetCreateSpaceResponse) Rows() [][]string {
	return [][]string{{
		truncateID(r.Name, 40),
		r.MeetingUri,
		r.MeetingCode,
	}}
}

func (r *MeetCreateSpaceResponse) EmptyMessage() string {
	return "No space created"
}

package types

const (
	maxMessageTextLength = 50
	truncatedTextLength  = 47
)

// ChatSpace represents a Google Chat space
type ChatSpace struct {
	ID                  string `json:"id"`
	Name                string `json:"name,omitempty"`
	Type                string `json:"type"`
	DisplayName         string `json:"displayName,omitempty"`
	Threaded            bool   `json:"threaded,omitempty"`
	ExternalUserAllowed bool   `json:"externalUserAllowed,omitempty"`
	SpaceHistoryState   string `json:"spaceHistoryState,omitempty"`
	CreateTime          string `json:"createTime,omitempty"`
}

func (s *ChatSpace) Headers() []string {
	return []string{"Space ID", "Name", "Type", "Display Name"}
}

func (s *ChatSpace) Rows() [][]string {
	return [][]string{{
		truncateID(s.ID, 30),
		s.Name,
		s.Type,
		s.DisplayName,
	}}
}

func (s *ChatSpace) EmptyMessage() string {
	return "No space information available"
}

// ChatMessage represents a message in Google Chat
type ChatMessage struct {
	ID            string `json:"id"`
	SpaceID       string `json:"spaceId,omitempty"`
	ThreadID      string `json:"threadId,omitempty"`
	Text          string `json:"text,omitempty"`
	FormattedText string `json:"formattedText,omitempty"`
	SenderName    string `json:"senderName,omitempty"`
	SenderEmail   string `json:"senderEmail,omitempty"`
	CreateTime    string `json:"createTime,omitempty"`
	UpdateTime    string `json:"updateTime,omitempty"`
}

func (m *ChatMessage) Headers() []string {
	return []string{"Message ID", "Sender", "Text", "Create Time"}
}

func (m *ChatMessage) Rows() [][]string {
	text := m.Text
	if len(text) > maxMessageTextLength {
		text = text[:truncatedTextLength] + "..."
	}
	return [][]string{{
		truncateID(m.ID, 20),
		m.SenderName,
		text,
		m.CreateTime,
	}}
}

func (m *ChatMessage) EmptyMessage() string {
	return "No message information available"
}

// ChatMember represents a member of a Google Chat space
type ChatMember struct {
	ID       string `json:"id"`
	SpaceID  string `json:"spaceId,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	State    string `json:"state,omitempty"`
	JoinTime string `json:"joinTime,omitempty"`
}

func (m *ChatMember) Headers() []string {
	return []string{"Member ID", "Name", "Email", "Role", "State"}
}

func (m *ChatMember) Rows() [][]string {
	return [][]string{{
		truncateID(m.ID, 20),
		m.Name,
		m.Email,
		m.Role,
		m.State,
	}}
}

func (m *ChatMember) EmptyMessage() string {
	return "No member information available"
}

// ChatThread represents a thread in a Google Chat space
type ChatThread struct {
	ID         string `json:"id"`
	SpaceID    string `json:"spaceId,omitempty"`
	Name       string `json:"name,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

func (t *ChatThread) Headers() []string {
	return []string{"Thread ID", "Name", "Create Time"}
}

func (t *ChatThread) Rows() [][]string {
	return [][]string{{
		truncateID(t.ID, 20),
		t.Name,
		t.CreateTime,
	}}
}

func (t *ChatThread) EmptyMessage() string {
	return "No thread information available"
}

// ChatSpacesListResponse represents a list of spaces response
type ChatSpacesListResponse struct {
	Spaces        []ChatSpace `json:"spaces"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

func (r *ChatSpacesListResponse) Headers() []string {
	return []string{"Space ID", "Name", "Type", "Display Name"}
}

func (r *ChatSpacesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Spaces))
	for i, space := range r.Spaces {
		rows[i] = []string{
			truncateID(space.ID, 30),
			space.Name,
			space.Type,
			space.DisplayName,
		}
	}
	return rows
}

func (r *ChatSpacesListResponse) EmptyMessage() string {
	return "No spaces found"
}

// ChatMessagesListResponse represents a list of messages response
type ChatMessagesListResponse struct {
	Messages      []ChatMessage `json:"messages"`
	NextPageToken string        `json:"nextPageToken,omitempty"`
}

func (r *ChatMessagesListResponse) Headers() []string {
	return []string{"Message ID", "Sender", "Text", "Create Time"}
}

func (r *ChatMessagesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Messages))
	for i, msg := range r.Messages {
		text := msg.Text
		if len(text) > maxMessageTextLength {
			text = text[:truncatedTextLength] + "..."
		}
		rows[i] = []string{
			truncateID(msg.ID, 20),
			msg.SenderName,
			text,
			msg.CreateTime,
		}
	}
	return rows
}

func (r *ChatMessagesListResponse) EmptyMessage() string {
	return "No messages found"
}

// ChatMembersListResponse represents a list of members response
type ChatMembersListResponse struct {
	Members       []ChatMember `json:"members"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

func (r *ChatMembersListResponse) Headers() []string {
	return []string{"Member ID", "Name", "Email", "Role", "State"}
}

func (r *ChatMembersListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Members))
	for i, member := range r.Members {
		rows[i] = []string{
			truncateID(member.ID, 20),
			member.Name,
			member.Email,
			member.Role,
			member.State,
		}
	}
	return rows
}

func (r *ChatMembersListResponse) EmptyMessage() string {
	return "No members found"
}

// ChatCreateMessageResponse represents a create message response
type ChatCreateMessageResponse struct {
	ID         string `json:"id"`
	SpaceID    string `json:"spaceId,omitempty"`
	ThreadID   string `json:"threadId,omitempty"`
	Text       string `json:"text,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

func (r *ChatCreateMessageResponse) Headers() []string {
	return []string{"Message ID", "Space ID", "Thread ID", "Create Time"}
}

func (r *ChatCreateMessageResponse) Rows() [][]string {
	return [][]string{{
		truncateID(r.ID, 20),
		truncateID(r.SpaceID, 20),
		truncateID(r.ThreadID, 20),
		r.CreateTime,
	}}
}

func (r *ChatCreateMessageResponse) EmptyMessage() string {
	return "No message created"
}

// ChatCreateSpaceResponse represents a create space response
type ChatCreateSpaceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

func (r *ChatCreateSpaceResponse) Headers() []string {
	return []string{"Space ID", "Name", "Type"}
}

func (r *ChatCreateSpaceResponse) Rows() [][]string {
	return [][]string{{
		truncateID(r.ID, 30),
		r.Name,
		r.Type,
	}}
}

func (r *ChatCreateSpaceResponse) EmptyMessage() string {
	return "No space created"
}

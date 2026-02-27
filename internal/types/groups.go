package types

type MemberRole struct {
	Name string `json:"name"`
}

type CloudIdentityGroup struct {
	Name              string            `json:"name"`
	GroupKeyID        string            `json:"groupKeyId"`
	GroupKeyNamespace string            `json:"groupKeyNamespace,omitempty"`
	DisplayName       string            `json:"displayName"`
	Description       string            `json:"description,omitempty"`
	CreateTime        string            `json:"createTime,omitempty"`
	UpdateTime        string            `json:"updateTime,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

func (g *CloudIdentityGroup) Headers() []string {
	return []string{"Name", "Display Name", "Email", "Created"}
}

func (g *CloudIdentityGroup) Rows() [][]string {
	return [][]string{{
		g.Name,
		g.DisplayName,
		g.GroupKeyID,
		g.CreateTime,
	}}
}

func (g *CloudIdentityGroup) EmptyMessage() string {
	return "No group found"
}

type CloudIdentityGroupList struct {
	Groups []CloudIdentityGroup `json:"groups"`
}

func (l *CloudIdentityGroupList) Headers() []string {
	return []string{"Name", "Display Name", "Email", "Created"}
}

func (l *CloudIdentityGroupList) Rows() [][]string {
	rows := make([][]string, len(l.Groups))
	for i, group := range l.Groups {
		rows[i] = []string{
			group.Name,
			group.DisplayName,
			group.GroupKeyID,
			group.CreateTime,
		}
	}
	return rows
}

func (l *CloudIdentityGroupList) EmptyMessage() string {
	return "No groups found"
}

type CloudIdentityMember struct {
	Name                        string       `json:"name"`
	PreferredMemberKeyID        string       `json:"preferredMemberKeyId"`
	PreferredMemberKeyNamespace string       `json:"preferredMemberKeyNamespace,omitempty"`
	Roles                       []MemberRole `json:"roles,omitempty"`
	CreateTime                  string       `json:"createTime,omitempty"`
}

func (m *CloudIdentityMember) Headers() []string {
	return []string{"Member", "Email", "Role", "Joined"}
}

func (m *CloudIdentityMember) Rows() [][]string {
	role := ""
	if len(m.Roles) > 0 {
		role = m.Roles[0].Name
	}
	return [][]string{{
		m.Name,
		m.PreferredMemberKeyID,
		role,
		m.CreateTime,
	}}
}

func (m *CloudIdentityMember) EmptyMessage() string {
	return "No member found"
}

type CloudIdentityMemberList struct {
	Members []CloudIdentityMember `json:"members"`
}

func (l *CloudIdentityMemberList) Headers() []string {
	return []string{"Member", "Email", "Role", "Joined"}
}

func (l *CloudIdentityMemberList) Rows() [][]string {
	rows := make([][]string, len(l.Members))
	for i, member := range l.Members {
		role := ""
		if len(member.Roles) > 0 {
			role = member.Roles[0].Name
		}
		rows[i] = []string{
			member.Name,
			member.PreferredMemberKeyID,
			role,
			member.CreateTime,
		}
	}
	return rows
}

func (l *CloudIdentityMemberList) EmptyMessage() string {
	return "No members found"
}

package api

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type ServiceAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Role struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionItem struct {
	ID          string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Acl struct {
	Resource     string `json:"resource"`
	Service      string `json:"service"`
	PermissionId string `json:"permissionId"`
	ResourceId   string `json:"resourceId"`
}

type AccessRights struct {
	IdentityUuid   string `json:"identityUuid"`
	IdentityType   string `json:"identityType"`
	Domain         string `json:"domain"`
	ResourceType   string `json:"resourceType"`
	ResourceUuid   string `json:"resourceUuid"`
	PathType       string `json:"pathType"`
	AccessTypeName string `json:"accessTypeName"`
	AccessTypeId   string `json:"accessTypeId"`
}

type CSVRow struct {
	EntityID        string
	EntityName      string
	EntityEmail     string
	Domain          string
	ResourceType    string
	ItemDescription string
	ItemType        string
	ItemId          string
	ItemName        string
	ItemAction      string
}

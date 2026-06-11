package api

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

var serviceResourceCombos = []struct {
	Service  string
	Resource string
}{
	{"iam", "user"},
	{"iam", "serviceAccount"},
	{"iam", "role"},
	{"iam", "permission"},
	{"postgresql", "backup"},
	{"postgresql", "cluster"},
	{"kubernetes", "cluster"},
	{"openshift", "cluster"},
	{"objectstorage", "bucket"},
	{"connectivity", "bridge"},
	{"connectivity", "directlink"},
	{"connectivity", "vpnconnection"},
}

func sanitizeField(s string) string {
	t := strings.TrimSpace(s)
	tl := strings.ToLower(t)
	if tl == "uuid.nil" || tl == "00000000-0000-0000-0000-000000000000" {
		return ""
	}
	return t
}

func (c *Client) ProcessUsers(ctx context.Context, logChan chan<- LogMessage) ([]CSVRow, error) {
	users, err := c.GetUsers(ctx, logChan)
	if err != nil {
		return nil, err
	}

	permsMap, err := c.GetPermissions(ctx)
	if err != nil {
		logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to fetch permissions map: %v", err)}
	}

	var allRows []CSVRow
	for i, u := range users {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		logChan <- LogMessage{Level: "info", Message: fmt.Sprintf("Processing user %d/%d: %s", i+1, len(users), u.Email)}

		roles, permissions, err := c.GetUserIamPolicies(ctx, u.ID)
		if err != nil {
			logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to get policies for %s: %v", u.Email, err)}
		}

		var aclsAll []Acl
		for _, combo := range serviceResourceCombos {
			acls, err := c.GetUserAcls(ctx, u.ID, combo.Service, combo.Resource)
			if err != nil {
				logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to get ACLs %s/%s for %s: %v", combo.Service, combo.Resource, u.Email, err)}
				continue
			}
			aclsAll = append(aclsAll, acls...)
		}

		for _, permID := range permissions {
			var permName, permDesc string
			if p, ok := permsMap[permID]; ok {
				permName = p.Name
				permDesc = p.Description
			} else {
				if p, err := c.GetPermissionDetails(ctx, permID); err == nil {
					permName = p.Name
					permDesc = p.Description
				}
			}
			allRows = append(allRows, CSVRow{
				EntityID:        u.ID,
				EntityName:      u.Email,
				ItemType:        "Permission",
				ItemId:          permID,
				ItemName:        permName,
				ItemDescription: permDesc,
				ItemAction:      extractActionName(permName),
			})
		}

		for _, roleID := range roles {
			var roleName, roleDesc string
			if r, err := c.GetRoleDetails(ctx, roleID); err == nil {
				roleName = r.Name
				roleDesc = r.Description
			}
			allRows = append(allRows, CSVRow{
				EntityID:        u.ID,
				EntityName:      u.Email,
				ItemType:        "Role",
				ItemId:          roleID,
				ItemName:        roleName,
				ItemDescription: roleDesc,
			})
		}

		for _, acl := range aclsAll {
			var permName string
			if p, ok := permsMap[acl.PermissionId]; ok {
				permName = p.Name
			} else {
				if p, err := c.GetPermissionDetails(ctx, acl.PermissionId); err == nil {
					permName = p.Name
				}
			}
			allRows = append(allRows, CSVRow{
				EntityID:        u.ID,
				EntityName:      u.Email,
				ItemType:        "ACL",
				ItemId:          acl.PermissionId,
				ItemName:        permName,
				ItemDescription: acl.ResourceId,
				ItemAction:      extractActionName(permName),
			})
		}

		if len(roles) == 0 && len(permissions) == 0 && len(aclsAll) == 0 {
			allRows = append(allRows, CSVRow{
				EntityID:   u.ID,
				EntityName: u.Email,
				ItemType:   "User",
				ItemName:   "No specific items found",
			})
		}
	}

	return allRows, nil
}

func (c *Client) ProcessServiceAccounts(ctx context.Context, logChan chan<- LogMessage) ([]CSVRow, error) {
	sas, err := c.GetServiceAccounts(ctx, logChan)
	if err != nil {
		return nil, err
	}

	permsMap, err := c.GetPermissions(ctx)
	if err != nil {
		logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to fetch permissions map: %v", err)}
	}

	var allRows []CSVRow
	for i, sa := range sas {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		logChan <- LogMessage{Level: "info", Message: fmt.Sprintf("Processing service account %d/%d: %s", i+1, len(sas), sa.Name)}

		roles, permissions, err := c.GetServiceAccountIamPolicies(ctx, sa.ID)
		if err != nil {
			logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to get policies for %s: %v", sa.Name, err)}
		}

		var aclsAll []Acl
		for _, combo := range serviceResourceCombos {
			acls, err := c.GetServiceAccountAcls(ctx, sa.ID, combo.Service, combo.Resource)
			if err != nil {
				logChan <- LogMessage{Level: "warn", Message: fmt.Sprintf("Failed to get ACLs %s/%s for %s: %v", combo.Service, combo.Resource, sa.Name, err)}
				continue
			}
			aclsAll = append(aclsAll, acls...)
		}

		for _, permID := range permissions {
			var permName, permDesc string
			if p, ok := permsMap[permID]; ok {
				permName = p.Name
				permDesc = p.Description
			} else {
				if p, err := c.GetPermissionDetails(ctx, permID); err == nil {
					permName = p.Name
					permDesc = p.Description
				}
			}
			allRows = append(allRows, CSVRow{
				EntityID:        sa.ID,
				EntityName:      sa.Name,
				ItemType:        "Permission",
				ItemId:          permID,
				ItemName:        permName,
				ItemDescription: permDesc,
				ItemAction:      extractActionName(permName),
			})
		}

		for _, roleID := range roles {
			var roleName, roleDesc string
			if r, err := c.GetRoleDetails(ctx, roleID); err == nil {
				roleName = r.Name
				roleDesc = r.Description
			}
			allRows = append(allRows, CSVRow{
				EntityID:        sa.ID,
				EntityName:      sa.Name,
				ItemType:        "Role",
				ItemId:          roleID,
				ItemName:        roleName,
				ItemDescription: roleDesc,
			})
		}

		for _, acl := range aclsAll {
			var permName string
			if p, ok := permsMap[acl.PermissionId]; ok {
				permName = p.Name
			} else {
				if p, err := c.GetPermissionDetails(ctx, acl.PermissionId); err == nil {
					permName = p.Name
				}
			}
			allRows = append(allRows, CSVRow{
				EntityID:        sa.ID,
				EntityName:      sa.Name,
				ItemType:        "ACL",
				ItemId:          acl.PermissionId,
				ItemName:        permName,
				ItemDescription: acl.ResourceId,
				ItemAction:      extractActionName(permName),
			})
		}

		if len(roles) == 0 && len(permissions) == 0 && len(aclsAll) == 0 {
			allRows = append(allRows, CSVRow{
				EntityID:   sa.ID,
				EntityName: sa.Name,
				ItemType:   "ServiceAccount",
				ItemName:   "No specific items found",
			})
		}
	}

	return allRows, nil
}

func (c *Client) ProcessByResource(ctx context.Context, resourceID string, logChan chan<- LogMessage) ([]CSVRow, error) {
	items, err := c.GetIdentityAccess(ctx, resourceID, logChan)
	if err != nil {
		return nil, err
	}

	nameToID, _ := c.GetPermissionsNameToIDMap(ctx)
	c.loadIdentityNames(ctx, logChan)

	sort.SliceStable(items, func(i, j int) bool {
		rankIdentityType := func(identityType string) int {
			switch identityType {
			case "user":
				return 0
			case "serviceAccount":
				return 1
			default:
				return 2
			}
		}
		rankPathType := func(pathType string) int {
			pt := strings.ToLower(pathType)
			switch pt {
			case "permission":
				return 0
			case "role":
				return 1
			case "acl":
				return 2
			default:
				return 3
			}
		}

		identityRankI := rankIdentityType(items[i].IdentityType)
		identityRankJ := rankIdentityType(items[j].IdentityType)
		if identityRankI != identityRankJ {
			return identityRankI < identityRankJ
		}
		if items[i].IdentityUuid != items[j].IdentityUuid {
			return items[i].IdentityUuid < items[j].IdentityUuid
		}
		pathRankI := rankPathType(items[i].PathType)
		pathRankJ := rankPathType(items[j].PathType)
		if pathRankI != pathRankJ {
			return pathRankI < pathRankJ
		}
		return false
	})

	return c.formatRowsWithIdentity(items, nameToID), nil
}

func (c *Client) ProcessByIdentity(ctx context.Context, identityID string, logChan chan<- LogMessage) ([]CSVRow, error) {
	items, err := c.GetResourcesByIdentity(ctx, identityID, logChan)
	if err != nil {
		return nil, err
	}

	nameToID, _ := c.GetPermissionsNameToIDMap(ctx)
	c.loadIdentityNames(ctx, logChan)

	sort.SliceStable(items, func(i, j int) bool {
		rankPathType := func(pathType string) int {
			pt := strings.ToLower(pathType)
			switch pt {
			case "permission":
				return 0
			case "role":
				return 1
			case "acl":
				return 2
			default:
				return 3
			}
		}

		pi := rankPathType(items[i].PathType)
		pj := rankPathType(items[j].PathType)
		if pi != pj {
			return pi < pj
		}
		if items[i].ResourceType != items[j].ResourceType {
			return items[i].ResourceType < items[j].ResourceType
		}
		if items[i].ResourceUuid != items[j].ResourceUuid {
			return items[i].ResourceUuid < items[j].ResourceUuid
		}
		return false
	})

	return c.formatRowsWithIdentity(items, nameToID), nil
}

func extractActionName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, ".", ":")
	parts := strings.Split(name, ":")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func (c *Client) formatRowsWithIdentity(items []AccessRights, nameToID map[string]string) []CSVRow {
	rows := make([]CSVRow, 0, len(items))
	for _, it := range items {
		email := c.getIdentityName(it.IdentityUuid)
		row := CSVRow{
			EntityID:        it.IdentityUuid,
			EntityName:      it.IdentityType,
			EntityEmail:     email,
			Domain:          it.Domain,
			ResourceType:    it.ResourceType,
			ItemDescription: it.ResourceUuid,
			ItemType:        it.PathType,
			ItemName:        it.AccessTypeName,
		}
		switch it.PathType {
		case "ACL":
			row.ItemId = ""
			row.ItemAction = extractActionName(it.AccessTypeName)
		case "Permission":
			row.ItemId = nameToID[it.AccessTypeName]
			row.ItemAction = extractActionName(it.AccessTypeName)
		default:
			row.ItemId = it.AccessTypeId
		}
		rows = append(rows, row)
	}
	return rows
}

func SanitizeRows(rows []CSVRow) []CSVRow {
	result := make([]CSVRow, len(rows))
	for i, row := range rows {
		result[i] = CSVRow{
			EntityID:        sanitizeField(row.EntityID),
			EntityName:      sanitizeField(row.EntityName),
			EntityEmail:     sanitizeField(row.EntityEmail),
			Domain:          sanitizeField(row.Domain),
			ResourceType:    sanitizeField(row.ResourceType),
			ItemDescription: sanitizeField(row.ItemDescription),
			ItemType:        sanitizeField(row.ItemType),
			ItemId:          sanitizeField(row.ItemId),
			ItemName:        sanitizeField(row.ItemName),
			ItemAction:      sanitizeField(row.ItemAction),
		}
	}
	return result
}

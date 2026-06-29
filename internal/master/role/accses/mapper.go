package role

import (
	"fmt"
	"time"

	rolepPermission "service/internal/master/role/permission"
)

func mapRequestToEntity(req RolPagesRequest) *RolPages {
	return &RolPages{
		Name:   req.Name,
		Icon:   req.Icon,
		URL:    req.URL,
		Level:  req.Level,
		Sort:   req.Sort,
		Parent: req.Parent,
		Active: req.Active,
	}
}

func mapEntityToResponse(e *RolPages) *RolPagesResponse {
	return &RolPagesResponse{
		Id:     e.Id,
		Name:   e.Name,
		Icon:   e.Icon,
		URL:    e.URL,
		Level:  e.Level,
		Sort:   e.Sort,
		Parent: e.Parent,
		Active: e.Active,
	}
}

// MapPermissionToRoleAccessResponse converts RolPermission to RoleAccessResponse structure
func MapPermissionToRoleAccessResponse(roles []string, permissions []*rolepPermission.RolPermission, pages []*RolPages) *RoleAccessResponse {
	response := &RoleAccessResponse{
		Roles:  roles,
		Access: []MenuItem{},
	}

	// Map permissions to create menu items based on pages
	// Filter pages yang memiliki permission read = true
	accessiblePages := []*RolPages{}
	for i, permission := range permissions {
		if permission.Read && i < len(pages) {
			accessiblePages = append(accessiblePages, pages[i])
		}
	}

	// Group pages by parent ID
	parentMap := make(map[int64][]*RolPages)
	var rootPages []*RolPages

	for _, page := range accessiblePages {
		if page.Parent == nil {
			rootPages = append(rootPages, page)
		} else {
			parentMap[*page.Parent] = append(parentMap[*page.Parent], page)
		}
	}

	// Build menu hierarchy
	for _, rootPage := range rootPages {
		menuItem := MenuItem{
			ID:    generateUUID(),
			Name:  rootPage.Name,
			Icon:  rootPage.Icon,
			URL:   rootPage.URL,
			Level: fmt.Sprintf("%d", rootPage.Level),
		}

		// Add children if any
		if children, exists := parentMap[rootPage.Id]; exists {
			for _, child := range children {
				menuChild := MenuChild{
					Name:  child.Name,
					Icon:  child.Icon,
					URL:   child.URL,
					Level: fmt.Sprintf("%d", child.Level),
				}
				menuItem.Children = append(menuItem.Children, menuChild)
			}
		}

		response.Access = append(response.Access, menuItem)
	}

	return response
}

// Helper function to generate UUID (you may want to use a proper UUID library)
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

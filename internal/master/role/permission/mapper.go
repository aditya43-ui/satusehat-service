package permission

// mapRequestToEntity converts RolPermissionRequest to *RolPermission entity
func mapRequestToEntity(req RolPermissionRequest) *RolPermission {
	return &RolPermission{
		Create:        req.Create,
		Read:          req.Read,
		Update:        req.Update,
		Disable:       req.Disable,
		Delete:        req.Delete,
		Active:        req.Active,
		FkRolPagesId:  req.FkRolPagesId,
		RoleKeycloak:  req.RoleKeycloak,
		GroupKeycloak: req.GroupKeycloak,
	}
}

// mapEntityToResponse converts *RolPermission entity to *RolPermissionResponse DTO
func mapEntityToResponse(e *RolPermission) *RolPermissionResponse {
	return &RolPermissionResponse{
		Id:             e.Id,
		Create:         e.Create,
		Read:           e.Read,
		Update:         e.Update,
		Disable:        e.Disable,
		Delete:         e.Delete,
		Active:         e.Active,
		FkRolPagesId:   e.FkRolPagesId,
		RoleMasterName: e.RoleMasterName,
		RoleKeycloak:   e.RoleKeycloak,
		GroupKeycloak:  e.GroupKeycloak,
	}
}

package master

// mapRequestToEntity converts RoleMasterRequest to *RoleMaster entity
func mapRequestToEntity(req RoleMasterRequest) *RoleMaster {
	entity := &RoleMaster{
		Name:      req.Name,
		Active:    req.Active,
		Slug:      req.Slug,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}

	if req.Id != nil {
		entity.Id = *req.Id
	}
	return entity
}

// mapEntityToResponse converts *RoleMaster entity to *RoleMasterResponse DTO
func mapEntityToResponse(e *RoleMaster) *RoleMasterResponse {
	return &RoleMasterResponse{
		Id:        e.Id,
		Name:      e.Name,
		Active:    e.Active,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		Slug:      e.Slug,
	}
}

package pages

// mapRequestToEntity converts RolPagesRequest to *RolPages entity
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

// mapEntityToResponse converts *RolPages entity to *RolPagesResponse DTO
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

// mapEntityToTreeResponse converts *RolPages entity to *RolPagesTreeResponse DTO
func mapEntityToTreeResponse(e *RolPages) *RolPagesTreeResponse {
	return &RolPagesTreeResponse{
		Id:     e.Id,
		Name:   e.Name,
		Icon:   e.Icon,
		URL:    e.URL,
		Level:  e.Level,
		Sort:   e.Sort,
		Active: e.Active,
	}
}

// buildTreeHierarchy membangun struktur hirarki dari flat list
func buildTreeHierarchy(pages []*RolPages) []*RolPagesTreeResponse {
	pageMap := make(map[int64]*RolPagesTreeResponse)
	var roots []*RolPagesTreeResponse

	// Convert semua pages ke TreeResponse dan buat map
	for _, page := range pages {
		treeResponse := mapEntityToTreeResponse(page)
		pageMap[page.Id] = treeResponse
	}

	// Bangun relasi parent-child
	for _, page := range pages {
		if page.Parent == nil {
			// Ini adalah root node
			roots = append(roots, pageMap[page.Id])
		} else {
			// Ini adalah child node
			if parent, exists := pageMap[*page.Parent]; exists {
				parent.Children = append(parent.Children, pageMap[page.Id])
			}
		}
	}

	// Sort children berdasarkan sort order
	for _, page := range pageMap {
		if len(page.Children) > 0 {
			sortChildren(page.Children)
		}
	}

	// Sort root nodes berdasarkan sort order
	sortChildren(roots)

	return roots
}

// sortChildren mengurutkan children berdasarkan field Sort
func sortChildren(children []*RolPagesTreeResponse) {
	for i := 0; i < len(children); i++ {
		for j := i + 1; j < len(children); j++ {
			if children[i].Sort > children[j].Sort {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

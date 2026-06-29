package permission

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"service/pkg/errors"
	"service/pkg/logger"
	"service/pkg/utils/query"
	"strings"
	"time"

	"gorm.io/gorm"

	// Import pages entity untuk referensi
	"service/internal/infrastructure/cache"
	rolePages "service/internal/master/role/pages"
)

type Service interface {
	GetList(ctx context.Context, limit, offset int, sorts []string, activeFilter *bool) (map[string]interface{}, error)
	GetDetail(ctx context.Context, id int64) (*RolPermissionResponse, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []string, limit, offset int) (map[string]interface{}, error)
	Create(ctx context.Context, req RolPermissionRequest) (*RolPermissionResponse, error)
	Update(ctx context.Context, id int64, req RolPermissionRequest) (*RolPermissionResponse, error)
	Delete(ctx context.Context, id int64) error
	GetByRoleAndGroup(ctx context.Context, roleKeycloak string, groupKeycloak []string, fkRolPagesId int64) (*RolPermissionResponse, error)
	GetByRoleAndPages(ctx context.Context, roleKeycloak string, fkRolPagesIds []int64) ([]*RolPermissionResponse, error)
	GetRolePermissionTree(ctx context.Context, roleKeycloak string, groupKeycloak []string, activeOnly *bool) (*RolePermissionTreeResponse, error)
	GetRolePermissionTreeRole(ctx context.Context, roleKeycloak string, activeOnly *bool) (*RolePermissionTreeResponse, error)
	GetRolePagesList(ctx context.Context, roleKeycloak string, limit, offset int, activeOnly *bool) (map[string]interface{}, error)
}

type service struct {
	cmdRepo   CommandRepository
	queryRepo QueryRepository
	cache     *cache.Manager
}

func NewService(cmdRepo CommandRepository, queryRepo QueryRepository, pagesRepo rolePages.QueryRepository, cacheManager *cache.Manager) Service {
	return &service{cmdRepo: cmdRepo, queryRepo: queryRepo, cache: cacheManager}
}

func (s *service) GetList(ctx context.Context, limit, offset int, sorts []string, activeFilter *bool) (map[string]interface{}, error) {
	filters := map[string]interface{}{}
	if activeFilter != nil {
		filters["active"] = *activeFilter
	}

	// Gunakan kapabilitas method Search agar filter dynamic & cache ikut ter-handle dengan efisien
	return s.Search(ctx, filters, sorts, limit, offset)
}

func (s *service) GetDetail(ctx context.Context, id int64) (*RolPermissionResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	entity, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RolPermission detail").Cause(err).Build()
	}
	if entity == nil {
		return nil, errors.NotFoundError().Message("RolPermission not found").Metadata("id", id).Build()
	}

	return mapEntityToResponse(entity), nil
}

func (s *service) Search(ctx context.Context, filters map[string]interface{}, sorts []string, limit, offset int) (map[string]interface{}, error) {
	if limit == 0 {
		limit = 10
	} else if limit < -1 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// --- 1. Implementasi Caching (Check) ---
	filterBytes, _ := json.Marshal(filters)
	sortBytes, _ := json.Marshal(sorts)
	// Membuat string gabungan dari filter, sort, dan pagination, lalu di-hash dengan SHA-256
	hashInput := fmt.Sprintf("%s|%s|%d|%d", string(filterBytes), string(sortBytes), limit, offset)
	hash := sha256.Sum256([]byte(hashInput))
	cacheKey := fmt.Sprintf("role_permission_search_v2:%s", hex.EncodeToString(hash[:]))

	if s.cache != nil {
		var strData string
		if err := s.cache.Get(ctx, cacheKey, &strData); err == nil && strData != "" {
			var cachedData struct {
				Data   []*RolPermissionResponse `json:"data"`
				Total  int64                    `json:"total"`
				Limit  int                      `json:"limit"`
				Offset int                      `json:"offset"`
			}
			if err := json.Unmarshal([]byte(strData), &cachedData); err == nil {
				logger.Default().Debug("Cache hit for Role Permission Search", logger.String("key", cacheKey))
				return map[string]interface{}{
					"data":   cachedData.Data,
					"total":  cachedData.Total,
					"limit":  cachedData.Limit,
					"offset": cachedData.Offset,
				}, nil
			}
		}
	}

	// Konversi sorts string ke SortField
	var sortFields []query.SortField
	for _, sort := range sorts {
		if sort == "" {
			continue
		}

		// Default ASC, tandai dengan - untuk DESC
		order := "ASC"
		column := sort
		if strings.HasPrefix(sort, "-") {
			order = "DESC"
			column = strings.TrimPrefix(sort, "-")
		} else if strings.HasPrefix(sort, "+") {
			column = strings.TrimPrefix(sort, "+")
		}

		// Validasi kolom yang diizinkan (lowercase untuk mapping)
		allowedColumns := map[string]bool{
			"id":               true,
			"create":           true,
			"read":             true,
			"update":           true,
			"disable":          true,
			"delete":           true,
			"active":           true,
			"fk_rol_pages_id":  true,
			"role_keycloak":    true,
			"group_keycloak":   true,
			"created_at":       true,
			"role_master_name": true,
		}

		if allowedColumns[column] {
			sortFields = append(sortFields, query.SortField{
				Column: column,
				Order:  order,
			})
		}
	}

	// Jika tidak ada sort yang valid, gunakan default
	if len(sortFields) == 0 {
		sortFields = []query.SortField{query.CreateAscSort("id")}
	}

	entities, total, err := s.queryRepo.Search(ctx, filters, sortFields, limit, offset)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to search RolPermissions").Cause(err).Build()
	}

	responses := make([]*RolPermissionResponse, len(entities))
	for i, entity := range entities {
		responses[i] = mapEntityToResponse(&entity)
	}

	responseMap := map[string]interface{}{
		"data": responses, "total": total, "limit": limit, "offset": offset,
	}

	// --- 2. Implementasi Caching (Set) ---
	if s.cache != nil {
		if bytes, err := json.Marshal(responseMap); err == nil {
			// TTL 5 menit untuk data search
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 5*time.Minute)
		}
	}

	return responseMap, nil
}

func (s *service) Create(ctx context.Context, req RolPermissionRequest) (*RolPermissionResponse, error) {
	// Validasi input
	if req.FkRolPagesId <= 0 {
		return nil, errors.NewValidationError().Message("Page ID is required").Metadata("fk_rol_pages_id", req.FkRolPagesId).Build()
	}
	if req.RoleKeycloak == "" && req.GroupKeycloak == "" {
		return nil, errors.NewValidationError().Message("Role or Group must be provided").Build()
	}

	entity := mapRequestToEntity(req)
	if err := s.cmdRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RolPermission with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to create RolPermission").Cause(err).Build()
	}
	// Re-fetch to get the complete created entity
	createdEntity, err := s.queryRepo.FindByID(ctx, entity.Id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve newly created RolPermission").Cause(err).Build()
	}

	// Invalidate cache setelah create permission baru
	if s.cache != nil {
		// Perhatian: Jika cache.Manager Anda tidak mendukung wildcard Delete secara native,
		// Anda mungkin perlu memanggil metode s.cache.DeleteByPrefix atau menghapus key spesifik.
		cacheKeyPattern := fmt.Sprintf("permission_tree:%s:*", req.RoleKeycloak)
		_ = s.cache.Delete(ctx, cacheKeyPattern)
		// Hapus juga cache list agar data terbaru langsung tampil
		_ = s.cache.Delete(ctx, "role_permission_list:*")
		// Hapus juga cache search
		_ = s.cache.Delete(ctx, "role_permission_search_v2:*")
	}

	return mapEntityToResponse(createdEntity), nil
}

func (s *service) Update(ctx context.Context, id int64, req RolPermissionRequest) (*RolPermissionResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah record ada
	existing, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RolPermission").Cause(err).Build()
	}
	if existing == nil {
		return nil, errors.NotFoundError().Message("RolPermission not found").Metadata("id", id).Build()
	}

	// Simpan role lama untuk invalidasi cache
	oldRoleKeycloak := existing.RoleKeycloak

	// Update entity dengan data dari request
	// existing.Create = req.Create
	// existing.Read = req.Read
	// existing.Update = req.Update
	// existing.Disable = req.Disable
	// existing.Delete = req.Delete
	// existing.Active = req.Active
	// existing.FkRolPagesId = req.FkRolPagesId
	// existing.RoleKeycloak = req.RoleKeycloak
	// existing.GroupKeycloak = req.GroupKeycloak

	if err := s.cmdRepo.Update(ctx, existing); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RolPermission with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to update RolPermission").Cause(err).Build()
	}

	// Invalidate cache
	if s.cache != nil {
		// Invalidate cache untuk role lama dan baru jika berbeda
		if oldRoleKeycloak != "" {
			_ = s.cache.Delete(ctx, fmt.Sprintf("permission_tree:%s:*", oldRoleKeycloak))
		}
		if existing.RoleKeycloak != "" && existing.RoleKeycloak != oldRoleKeycloak {
			_ = s.cache.Delete(ctx, fmt.Sprintf("permission_tree:%s:*", existing.RoleKeycloak))
		}
		// Invalidate cache list dan search
		_ = s.cache.Delete(ctx, "role_permission_list:*")
		_ = s.cache.Delete(ctx, "role_permission_search_v2:*")
	}

	return mapEntityToResponse(existing), nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah record ada untuk mendapatkan roleKeycloak sebelum dihapus
	existing, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalError().Message("Failed to retrieve RolPermission before deletion").Cause(err).Build()
	}
	if existing == nil {
		// Jika sudah tidak ada, anggap operasi berhasil (idempotent)
		return nil
	}

	if err := s.cmdRepo.Delete(ctx, id); err != nil {
		return errors.InternalError().Message("Failed to delete RolPermission").Cause(err).Build()
	}

	// Invalidate cache
	if s.cache != nil {
		if existing.RoleKeycloak != "" {
			_ = s.cache.Delete(ctx, fmt.Sprintf("permission_tree:%s:*", existing.RoleKeycloak))
		}
		// Invalidate cache list dan search
		_ = s.cache.Delete(ctx, "role_permission_list:*")
		_ = s.cache.Delete(ctx, "role_permission_search_v2:*")
	}

	return nil
}

func (s *service) GetByRoleAndGroup(ctx context.Context, roleKeycloak string, groupKeycloak []string, fkRolPagesId int64) (*RolPermissionResponse, error) {
	if roleKeycloak == "" {
		return nil, errors.NewValidationError().Message("Role must be provided").Build()
	}
	if fkRolPagesId <= 0 {
		return nil, errors.NewValidationError().Message("Invalid page ID").Build()
	}

	entity, err := s.queryRepo.FindByRoleAndGroup(ctx, roleKeycloak, groupKeycloak)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve permission by role and group").Cause(err).Build()
	}
	// Cek jika tidak ada permission yang ditemukan
	if len(entity) == 0 {
		return nil, errors.NotFoundError().Message("Permission not found for role and group").Build()
	}

	// Untuk saat ini, kita kembalikan data pertama yang ditemukan.
	// Anda mungkin perlu logika tambahan jika ada lebih dari satu hasil.
	response := mapEntityToResponse(entity[0])
	return response, nil
}

// Helper function to build page access tree
func buildPageAccessTree(pages []rolePages.RolPages, permissionMap map[int64]*RolPermission) []*PageAccess {
	pageMap := make(map[int64]*PageAccess)
	var roots []*PageAccess

	// Convert all pages to PageAccess
	for i := 0; i < len(pages); i++ {
		page := &pages[i]
		access := &PageAccess{
			Id:     page.Id,
			Name:   page.Name,
			Icon:   page.Icon,
			URL:    page.URL,
			Level:  page.Level,
			Sort:   page.Sort,
			Active: page.Active,
		}

		// Add permission if exists
		if perm, exists := permissionMap[page.Id]; exists {
			access.Permission = &PermissionDetail{
				Create:  perm.Create,
				Read:    perm.Read,
				Update:  perm.Update,
				Delete:  perm.Delete,
				Disable: perm.Disable,
			}
			access.Group = perm.GroupKeycloak
		}

		pageMap[page.Id] = access
	}

	// Build parent-child relationships
	for i := range pages {
		page := &pages[i]
		if page.Parent == nil {
			roots = append(roots, pageMap[page.Id])
		} else {
			if parent, exists := pageMap[*page.Parent]; exists {
				parent.Children = append(parent.Children, pageMap[page.Id])
			}
		}
	}

	// Sort children by sort order
	for _, access := range pageMap {
		if len(access.Children) > 0 {
			sortPageAccessChildren(access.Children)
		}
	}

	// Sort roots by sort order
	sortPageAccessChildren(roots)

	return roots
}

func sortPageAccessChildren(children []*PageAccess) {
	for i := 0; i < len(children); i++ {
		for j := i + 1; j < len(children); j++ {
			if children[i].Sort > children[j].Sort {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

// convertPageAccessToAccessTree converts []*PageAccess to []*AccessTreeItem
func convertPageAccessToAccessTree(pageAccess []*PageAccess) []*AccessTreeItem {
	result := make([]*AccessTreeItem, len(pageAccess))
	for i, access := range pageAccess {
		result[i] = &AccessTreeItem{
			ID:       access.Id,
			Name:     access.Name,
			Icon:     access.Icon,
			URL:      access.URL,
			Group:    access.Group,
			Level:    int(access.Level),
			Sort:     int(access.Sort),
			Active:   access.Active,
			Children: convertPageAccessToAccessTree(access.Children),
		}

		// Convert permission if exists
		if access.Permission != nil {
			result[i].Permission = &PermissionResponse{
				Id:      access.Id,
				Create:  access.Permission.Create,
				Read:    access.Permission.Read,
				Update:  access.Permission.Update,
				Delete:  access.Permission.Delete,
				Disable: access.Permission.Disable,
			}
		}
	}
	return result
}

// buildPermissionTreeHierarchy builds permission tree hierarchy from pages and permissions
func buildPermissionTreeHierarchy(pages []rolePages.RolPages, permissionMap map[int64]*RolPermission) []*PermissionAccessItem {
	pageMap := make(map[int64]*PermissionAccessItem)
	var roots []*PermissionAccessItem

	// Convert all pages to PermissionAccessItem
	for i := 0; i < len(pages); i++ {
		page := &pages[i]
		access := &PermissionAccessItem{
			Id:     page.Id,
			Name:   page.Name,
			Icon:   page.Icon,
			URL:    page.URL,
			Level:  page.Level,
			Sort:   page.Sort,
			Active: page.Active,
		}

		// Add permission if exists
		if perm, exists := permissionMap[page.Id]; exists {
			access.Permission = &PermissionDetail{
				Create:  perm.Create,
				Read:    perm.Read,
				Update:  perm.Update,
				Delete:  perm.Delete,
				Disable: perm.Disable,
			}
		}

		pageMap[page.Id] = access
	}

	// Build parent-child relationships
	for i := 0; i < len(pages); i++ {
		page := &pages[i]
		if page.Parent == nil {
			roots = append(roots, pageMap[page.Id])
		} else {
			if parent, exists := pageMap[*page.Parent]; exists {
				parent.Children = append(parent.Children, pageMap[page.Id])
			}
		}
	}

	// Sort children by sort order
	for _, access := range pageMap {
		if len(access.Children) > 0 {
			sortPermissionAccessChildren(access.Children)
		}
	}

	// Sort roots by sort order
	sortPermissionAccessChildren(roots)

	return roots
}

func sortPermissionAccessChildren(children []*PermissionAccessItem) {
	for i := 0; i < len(children); i++ {
		for j := i + 1; j < len(children); j++ {
			if children[i].Sort > children[j].Sort {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

func (s *service) GetByRoleAndPages(ctx context.Context, roleKeycloak string, fkRolPagesIds []int64) ([]*RolPermissionResponse, error) {
	if roleKeycloak == "" {
		return nil, errors.NewValidationError().Message("Role must be provided").Build()
	}
	if len(fkRolPagesIds) == 0 {
		return nil, errors.NewValidationError().Message("Page IDs must be provided").Build()
	}

	entities, err := s.queryRepo.FindByRoleAndPages(ctx, roleKeycloak, fkRolPagesIds)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve permissions by role and pages").Cause(err).Build()
	}

	responses := make([]*RolPermissionResponse, len(entities))
	for i, entity := range entities {
		responses[i] = mapEntityToResponse(entity)
	}

	return responses, nil
}

func (s *service) GetRolePermissionTree(ctx context.Context, roleKeycloak string, groupKeycloak []string, activeOnly *bool) (*RolePermissionTreeResponse, error) {
	if roleKeycloak == "" && len(groupKeycloak) == 0 {
		return nil, errors.NewValidationError().Message("Role or group must be provided").Build()
	}

	// --- 1. Implementasi Caching (Check) ---
	var activeStr string
	if activeOnly != nil {
		activeStr = fmt.Sprintf("%v", *activeOnly)
	} else {
		activeStr = "all"
	}

	cacheKey := fmt.Sprintf("permission_tree:%s:%s:%s", roleKeycloak, strings.Join(groupKeycloak, "_"), activeStr)

	if s.cache != nil {
		var strData string
		if err := s.cache.Get(ctx, cacheKey, &strData); err == nil && strData != "" {
			var cachedResponse RolePermissionTreeResponse
			if err := json.Unmarshal([]byte(strData), &cachedResponse); err == nil {
				logger.Default().Debug("Cache hit for Role Permission Tree", logger.String("key", cacheKey))
				return &cachedResponse, nil
			}
		}
	}

	// Gunakan query JOIN yang lebih efisien
	permissionsWithPages, err := s.queryRepo.FindPermissionsWithPages(ctx, roleKeycloak, groupKeycloak, activeOnly)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve permissions with pages").Cause(err).Build()
	}

	// Jika tidak ada hasil, return empty
	if len(permissionsWithPages) == 0 {
		return &RolePermissionTreeResponse{
			Success: true,
			Data: RolePermissionData{
				Role: roleKeycloak,
				// Group:  groupKeycloak,
				Access: []*AccessTreeItem{},
			},
		}, nil
	}

	// Konversi hasil JOIN ke struct pages dan permissions
	var pages []rolePages.RolPages
	pageMap := make(map[int64]bool)
	permissionMap := make(map[int64]*RolPermission)

	for _, permPage := range permissionsWithPages {
		// Buat atau dapatkan page
		if !pageMap[permPage.FkRolPagesId] {
			page := rolePages.RolPages{
				Id:     permPage.FkRolPagesId,
				Name:   permPage.PageName,
				URL:    permPage.PageUrl,
				Level:  int16(permPage.PageLevel),
				Sort:   int16(permPage.PageSort),
				Active: permPage.PageActive,
				Icon:   permPage.PageIcon,
			}
			if permPage.PageParentId != nil {
				page.Parent = permPage.PageParentId
			}
			pages = append(pages, page)
			pageMap[permPage.FkRolPagesId] = true
		}

		// Buat permission
		permission := &RolPermission{
			Id:            permPage.Id,
			Create:        permPage.Create,
			Read:          permPage.Read,
			Update:        permPage.Update,
			Disable:       permPage.Disable,
			Delete:        permPage.Delete,
			Active:        permPage.Active,
			FkRolPagesId:  permPage.FkRolPagesId,
			RoleKeycloak:  permPage.RoleKeycloak,
			GroupKeycloak: permPage.GroupKeycloak,
			CreatedAt:     permPage.CreatedAt,
			UpdatedAt:     permPage.UpdatedAt,
		}
		permissionMap[permPage.FkRolPagesId] = permission
	}

	// Build tree dengan pages yang sudah diurutkan
	access := buildPageAccessTree(pages, permissionMap)

	// Convert to proper response format
	accessTree := convertPageAccessToAccessTree(access)

	// Ambil group_keycloak yang sebenarnya dari data permission (jika ada)
	var actualGroup []string
	if len(permissionsWithPages) > 0 {
		// Kumpulkan semua group_keycloak yang unik
		groupMap := make(map[string]bool)
		for _, permPage := range permissionsWithPages {
			if permPage.GroupKeycloak != "" {
				groupMap[permPage.GroupKeycloak] = true
			}
		}
		// Konversi map ke slice
		for group := range groupMap {
			actualGroup = append(actualGroup, group)
		}
	}

	response := &RolePermissionTreeResponse{
		Success: true,
		Data: RolePermissionData{
			Role: roleKeycloak,
			// Group:  actualGroup,
			Access: accessTree,
		},
	}

	// --- 2. Implementasi Caching (Set) ---
	if s.cache != nil {
		if bytes, err := json.Marshal(response); err == nil {
			// Simpan ke cache dengan TTL 1 jam (Dapat disesuaikan)
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 1*time.Hour)
		}
	}

	return response, nil
}

func (s *service) GetRolePagesList(ctx context.Context, roleKeycloak string, limit, offset int, activeOnly *bool) (map[string]interface{}, error) {
	// 1. Inisiasi Limit & Offset standar
	if limit == 0 {
		limit = 10
	} else if limit < -1 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 2. Fetch dataAll: semua role pages aktif beserta permission mapping (query FindActivePagesWithPermission)
	isActive := true
	if activeOnly != nil {
		isActive = *activeOnly
	}
	dataAllRecords, err := s.queryRepo.FindActivePagesWithPermission(ctx, isActive)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve active pages for dataAll").Cause(err).Build()
	}

	var allPages []rolePages.RolPages
	allPageMap := make(map[int64]bool)
	allPermissionMap := make(map[int64]*RolPermission)
	for _, permPage := range dataAllRecords {
		if !allPageMap[permPage.FkRolPagesId] {
			page := rolePages.RolPages{
				Id:     permPage.FkRolPagesId,
				Name:   permPage.PageName,
				URL:    permPage.PageUrl,
				Level:  int16(permPage.PageLevel),
				Sort:   int16(permPage.PageSort),
				Active: permPage.PageActive,
				Icon:   permPage.PageIcon,
			}
			if permPage.PageParentId != nil {
				page.Parent = permPage.PageParentId
			}
			allPages = append(allPages, page)
			allPageMap[permPage.FkRolPagesId] = true
		}
		allPermissionMap[permPage.FkRolPagesId] = &RolPermission{
			Id:            permPage.Id,
			Create:        permPage.Create,
			Read:          permPage.Read,
			Update:        permPage.Update,
			Delete:        permPage.Delete,
			Disable:       permPage.Disable,
			GroupKeycloak: permPage.GroupKeycloak,
			RoleKeycloak:  permPage.RoleKeycloak,
		}
	}
	accessAll := buildPageAccessTree(allPages, allPermissionMap)
	accessTreeAll := convertPageAccessToAccessTree(accessAll)

	// 3. Fetch seluruh role permission berserta info pages-nya (dengan filter roleKeycloak spesifik)
	permissionsWithPages, err := s.queryRepo.FindPermissionsWithRole(ctx, roleKeycloak, activeOnly)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve permissions with pages").Cause(err).Build()
	}

	var pages []rolePages.RolPages
	pageMap := make(map[int64]bool)
	permissionMap := make(map[int64]*RolPermission)

	for _, permPage := range permissionsWithPages {
		if !pageMap[permPage.FkRolPagesId] {
			page := rolePages.RolPages{
				Id:     permPage.FkRolPagesId,
				Name:   permPage.PageName,
				URL:    permPage.PageUrl,
				Level:  int16(permPage.PageLevel),
				Sort:   int16(permPage.PageSort),
				Active: permPage.PageActive,
				Icon:   permPage.PageIcon,
			}
			if permPage.PageParentId != nil {
				page.Parent = permPage.PageParentId
			}
			pages = append(pages, page)
			pageMap[permPage.FkRolPagesId] = true
		}

		permissionMap[permPage.FkRolPagesId] = &RolPermission{
			Id:            permPage.Id,
			Create:        permPage.Create,
			Read:          permPage.Read,
			Update:        permPage.Update,
			Delete:        permPage.Delete,
			Disable:       permPage.Disable,
			GroupKeycloak: permPage.GroupKeycloak,
			RoleKeycloak:  permPage.RoleKeycloak,
		}
	}

	// Build tree menggunakan helper yang telah ada
	access := buildPageAccessTree(pages, permissionMap)
	accessTree := convertPageAccessToAccessTree(access)

	data := []*RolePageAccessResponse{
		{
			RoleID:   roleKeycloak,
			DataRole: accessTree,
			DataAll:  accessTreeAll,
		},
	}

	return map[string]interface{}{
		"data":   data,
		"total":  int64(1),
		"limit":  limit,
		"offset": offset,
	}, nil
}

func (s *service) GetRolePermissionTreeRole(ctx context.Context, roleKeycloak string, activeOnly *bool) (*RolePermissionTreeResponse, error) {

	// --- 1. Implementasi Caching (Check) ---
	var activeStr string
	if activeOnly != nil {
		activeStr = fmt.Sprintf("%v", *activeOnly)
	} else {
		activeStr = "all"
	}

	cacheKey := fmt.Sprintf("permission_tree:%s:%s", roleKeycloak, activeStr)

	if s.cache != nil {
		var strData string
		if err := s.cache.Get(ctx, cacheKey, &strData); err == nil && strData != "" {
			var cachedResponse RolePermissionTreeResponse
			if err := json.Unmarshal([]byte(strData), &cachedResponse); err == nil {
				logger.Default().Debug("Cache hit for Role Permission Tree", logger.String("key", cacheKey))
				return &cachedResponse, nil
			}
		}
	}

	// Gunakan query JOIN yang lebih efisien
	// permissionsWithPages, err := s.queryRepo.FindPermissionsWithPages(ctx, roleKeycloak, groupKeycloak, activeOnly)
	permissionsWithPages, err := s.queryRepo.FindPermissionsWithRole(ctx, roleKeycloak, activeOnly)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve permissions with pages").Cause(err).Build()
	}

	// Jika tidak ada hasil, return empty
	if len(permissionsWithPages) == 0 {
		return &RolePermissionTreeResponse{
			Success: true,
			Data: RolePermissionData{
				Role: roleKeycloak,
				// Group:  groupKeycloak,
				Access: []*AccessTreeItem{},
			},
		}, nil
	}

	// Konversi hasil JOIN ke struct pages dan permissions
	var pages []rolePages.RolPages
	pageMap := make(map[int64]bool)
	permissionMap := make(map[int64]*RolPermission)

	for _, permPage := range permissionsWithPages {
		// Buat atau dapatkan page
		if !pageMap[permPage.FkRolPagesId] {
			page := rolePages.RolPages{
				Id:     permPage.FkRolPagesId,
				Name:   permPage.PageName,
				URL:    permPage.PageUrl,
				Level:  int16(permPage.PageLevel),
				Sort:   int16(permPage.PageSort),
				Active: permPage.PageActive,
				Icon:   permPage.PageIcon,
			}
			if permPage.PageParentId != nil {
				page.Parent = permPage.PageParentId
			}
			pages = append(pages, page)
			pageMap[permPage.FkRolPagesId] = true
		}

		// Buat permission
		permission := &RolPermission{
			Id:            permPage.Id,
			Create:        permPage.Create,
			Read:          permPage.Read,
			Update:        permPage.Update,
			Disable:       permPage.Disable,
			Delete:        permPage.Delete,
			Active:        permPage.Active,
			FkRolPagesId:  permPage.FkRolPagesId,
			RoleKeycloak:  permPage.RoleKeycloak,
			GroupKeycloak: permPage.GroupKeycloak,
			CreatedAt:     permPage.CreatedAt,
			UpdatedAt:     permPage.UpdatedAt,
		}
		permissionMap[permPage.FkRolPagesId] = permission
	}

	// Build tree dengan pages yang sudah diurutkan
	access := buildPageAccessTree(pages, permissionMap)

	// Convert to proper response format
	accessTree := convertPageAccessToAccessTree(access)

	// Ambil group_keycloak yang sebenarnya dari data permission (jika ada)
	var actualGroup []string
	if len(permissionsWithPages) > 0 {
		// Kumpulkan semua group_keycloak yang unik
		groupMap := make(map[string]bool)
		for _, permPage := range permissionsWithPages {
			if permPage.GroupKeycloak != "" {
				groupMap[permPage.GroupKeycloak] = true
			}
		}
		// Konversi map ke slice
		for group := range groupMap {
			actualGroup = append(actualGroup, group)
		}
	}

	response := &RolePermissionTreeResponse{
		Success: true,
		Data: RolePermissionData{
			Role: roleKeycloak,
			// Group:  actualGroup,
			Access: accessTree,
		},
	}

	// --- 2. Implementasi Caching (Set) ---
	if s.cache != nil {
		if bytes, err := json.Marshal(response); err == nil {
			// Simpan ke cache dengan TTL 1 jam (Dapat disesuaikan)
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 1*time.Hour)
		}
	}

	return response, nil
}

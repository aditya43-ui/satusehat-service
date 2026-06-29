package pages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"service/internal/infrastructure/cache"
	"service/pkg/errors"
	"service/pkg/logger"
	"service/pkg/utils/query"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	GetList(ctx context.Context, limit, offset int, sorts []string, activeFilter *bool) (map[string]interface{}, error)
	GetDetail(ctx context.Context, id int64) (*RolPagesResponse, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []string, limit, offset int) (map[string]interface{}, error)
	Create(ctx context.Context, req RolPagesRequest) (*RolPagesResponse, error)
	Update(ctx context.Context, id int64, req RolPagesRequest) (*RolPagesResponse, error)
	Delete(ctx context.Context, id int64) error
	GetTree(ctx context.Context, activeFilter *bool) ([]*RolPagesTreeResponse, error)
	GetTreeByLevel(ctx context.Context, level int16, activeFilter *bool) ([]*RolPagesTreeResponse, error)
}

type service struct {
	cmdRepo   CommandRepository
	queryRepo QueryRepository
	cache     *cache.Manager
}

func NewService(cmdRepo CommandRepository, queryRepo QueryRepository, cacheManager *cache.Manager) Service {
	return &service{cmdRepo: cmdRepo, queryRepo: queryRepo, cache: cacheManager}
}

func (s *service) GetList(ctx context.Context, limit, offset int, sorts []string, activeFilter *bool) (map[string]interface{}, error) {
	filters := map[string]interface{}{}
	if activeFilter != nil {
		filters["active"] = *activeFilter
	}
	return s.Search(ctx, filters, sorts, limit, offset)
}

func (s *service) GetDetail(ctx context.Context, id int64) (*RolPagesResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	entity, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RolPages detail").Cause(err).Build()
	}
	if entity == nil {
		return nil, errors.NotFoundError().Message("RolPages not found").Metadata("id", id).Build()
	}

	response := mapEntityToResponse(entity)

	// Ambil children jika ada
	children, err := s.queryRepo.GetChildren(ctx, id)
	if err == nil && len(children) > 0 {
		response.Children = make([]*RolPagesResponse, len(children))
		for i := range children {
			response.Children[i] = mapEntityToResponse(&children[i])
		}
	}

	return response, nil
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
	hashInput := fmt.Sprintf("%s|%s|%d|%d", string(filterBytes), string(sortBytes), limit, offset)
	hash := sha256.Sum256([]byte(hashInput))
	cacheKey := fmt.Sprintf("role_pages_search_v2:%s", hex.EncodeToString(hash[:]))

	if s.cache != nil {
		var strData string
		if err := s.cache.Get(ctx, cacheKey, &strData); err == nil && strData != "" {
			var cachedData struct {
				Data   []*RolPagesResponse `json:"data"`
				Total  int64               `json:"total"`
				Limit  int                 `json:"limit"`
				Offset int                 `json:"offset"`
			}
			if err := json.Unmarshal([]byte(strData), &cachedData); err == nil {
				logger.Default().Debug("Cache hit for Role Pages Search", logger.String("key", cacheKey))
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
			"id":         true,
			"name":       true,
			"level":      true,
			"sort":       true,
			"active":     true,
			"parent":     true,
			"icon":       true,
			"url":        true,
			"created_at": true,
			"updated_at": true,
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
		return nil, errors.InternalError().Message("Failed to search RolPagess").Cause(err).Build()
	}

	responses := make([]*RolPagesResponse, len(entities))
	for i := range entities {
		responses[i] = mapEntityToResponse(&entities[i])
	}

	responseMap := map[string]interface{}{
		"data": responses, "total": total, "limit": limit, "offset": offset,
	}

	// --- 2. Implementasi Caching (Set) ---
	if s.cache != nil {
		if bytes, err := json.Marshal(responseMap); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(bytes), 5*time.Minute)
		}
	}

	return responseMap, nil
}

func (s *service) Create(ctx context.Context, req RolPagesRequest) (*RolPagesResponse, error) {
	// TODO: Add validation here if needed
	entity := mapRequestToEntity(req)
	if err := s.cmdRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RolPages with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to create RolPages").Cause(err).Build()
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "role_pages_search_v2:*")
	}
	return mapEntityToResponse(entity), nil
}

func (s *service) GetTree(ctx context.Context, activeFilter *bool) ([]*RolPagesTreeResponse, error) {
	filters := map[string]interface{}{}
	if activeFilter != nil {
		filters["active"] = *activeFilter
	}

	pages, _, err := s.queryRepo.Search(ctx, filters, []query.SortField{query.CreateAscSort("sort"), query.CreateAscSort("id")}, -1, 0)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve pages").Cause(err).Build()
	}

	// Konversi []RolPages ke []*RolPages
	pagePointers := make([]*RolPages, len(pages))
	for i := range pages {
		pagePointers[i] = &pages[i]
	}

	return buildTreeHierarchy(pagePointers), nil
}

func (s *service) GetTreeByLevel(ctx context.Context, level int16, activeFilter *bool) ([]*RolPagesTreeResponse, error) {
	filters := map[string]interface{}{"level": level}
	if activeFilter != nil {
		filters["active"] = *activeFilter
	}

	pages, _, err := s.queryRepo.Search(ctx, filters, []query.SortField{query.CreateAscSort("sort"), query.CreateAscSort("id")}, -1, 0)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve pages by level").Cause(err).Build()
	}

	// Konversi []RolPages ke []*RolPages
	pagePointers := make([]*RolPages, len(pages))
	for i := range pages {
		pagePointers[i] = &pages[i]
	}

	return buildTreeHierarchy(pagePointers), nil
}
func (s *service) Update(ctx context.Context, id int64, req RolPagesRequest) (*RolPagesResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah record ada
	existing, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RolPages").Cause(err).Build()
	}
	if existing == nil {
		return nil, errors.NotFoundError().Message("RolPages not found").Metadata("id", id).Build()
	}

	// Validasi parent jika berubah dan tidak boleh circular reference
	if req.Parent != nil && *req.Parent != id {
		if *req.Parent == id {
			return nil, errors.NewValidationError().Message("Cannot set self as parent").Build()
		}

		parent, err := s.queryRepo.FindByID(ctx, *req.Parent)
		if err != nil {
			return nil, errors.InternalError().Message("Failed to validate parent").Cause(err).Build()
		}
		if parent == nil {
			return nil, errors.NotFoundError().Message("Parent page not found").Metadata("parent_id", *req.Parent).Build()
		}
	}

	// Update entity
	existing.Name = req.Name
	existing.Icon = req.Icon
	existing.URL = req.URL
	existing.Level = req.Level
	existing.Sort = req.Sort
	existing.Parent = req.Parent
	existing.Active = req.Active

	if err := s.cmdRepo.Update(ctx, existing); err != nil {
		return nil, errors.InternalError().Message("Failed to update RolPages").Cause(err).Build()
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "role_pages_search_v2:*")
	}

	return mapEntityToResponse(existing), nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah memiliki children
	children, err := s.queryRepo.GetChildren(ctx, id)
	if err != nil {
		return errors.InternalError().Message("Failed to check children").Cause(err).Build()
	}
	if len(children) > 0 {
		return errors.NewValidationError().Message("Cannot delete page with children").Metadata("children_count", len(children)).Build()
	}

	if err := s.cmdRepo.Delete(ctx, id); err != nil {
		return errors.InternalError().Message("Failed to delete RolPages").Cause(err).Build()
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "role_pages_search_v2:*")
	}

	return nil
}

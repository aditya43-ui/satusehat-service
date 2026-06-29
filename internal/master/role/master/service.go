package master

import (
	"context"
	"service/internal/infrastructure/cache"
	"service/pkg/errors"
	"service/pkg/utils/query"
	"strings"

	"gorm.io/gorm"
)

type Service interface {
	GetList(ctx context.Context, limit, offset int, sorts []string, activeFilter *bool) (map[string]interface{}, error)
	GetDetail(ctx context.Context, id int64) (*RoleMasterResponse, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []string, limit, offset int) (map[string]interface{}, error)
	Create(ctx context.Context, req RoleMasterRequest) (*RoleMasterResponse, error)
	Update(ctx context.Context, id int64, req RoleMasterRequest) (*RoleMasterResponse, error)
	Upsert(ctx context.Context, req RoleMasterRequest) (*RoleMasterResponse, error)
	Delete(ctx context.Context, id int64) error
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

func (s *service) GetDetail(ctx context.Context, id int64) (*RoleMasterResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	entity, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RoleMaster detail").Cause(err).Build()
	}
	if entity == nil {
		return nil, errors.NotFoundError().Message("RoleMaster not found").Metadata("id", id).Build()
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
			"active":     true,
			"created_at": true,
			"updated_at": true,
			"select":     true,
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
		return nil, errors.InternalError().Message("Failed to search RoleMasters").Cause(err).Build()
	}

	responses := make([]*RoleMasterResponse, len(entities))
	for i, entity := range entities {
		responses[i] = mapEntityToResponse(&entity)
	}

	return map[string]interface{}{
		"data": responses, "total": total, "limit": limit, "offset": offset,
	}, nil
}

func (s *service) Create(ctx context.Context, req RoleMasterRequest) (*RoleMasterResponse, error) {
	// TODO: Add validation here if needed
	entity := mapRequestToEntity(req)
	if err := s.cmdRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RoleMaster with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to create RoleMaster").Cause(err).Build()
	}
	// Re-fetch to get the complete created entity
	createdEntity, err := s.queryRepo.FindByID(ctx, entity.Id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve newly created RoleMaster").Cause(err).Build()
	}
	return mapEntityToResponse(createdEntity), nil
}

func (s *service) Update(ctx context.Context, id int64, req RoleMasterRequest) (*RoleMasterResponse, error) {
	if id <= 0 {
		return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah record ada
	existing, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve RoleMaster").Cause(err).Build()
	}
	if existing == nil {
		return nil, errors.NotFoundError().Message("RoleMaster not found").Metadata("id", id).Build()
	}

	// Update fields from request on the existing entity
	existing.Name = req.Name
	existing.Active = req.Active
	existing.Slug = req.Slug
	existing.CreatedAt = req.CreatedAt
	existing.UpdatedAt = req.UpdatedAt

	if err := s.cmdRepo.Update(ctx, existing); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RoleMaster with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to update RoleMaster").Cause(err).Build()
	}

	// TODO: Add cache invalidation logic here

	return mapEntityToResponse(existing), nil
}

func (s *service) Upsert(ctx context.Context, req RoleMasterRequest) (*RoleMasterResponse, error) {
	entity := mapRequestToEntity(req)

	// Lakukan proses Create dengan skenario OnConflict Do Update
	if err := s.cmdRepo.Upsert(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.AlreadyExistsError().Message("RoleMaster with this identifier already exists").Cause(err).Build()
		}
		return nil, errors.InternalError().Message("Failed to upsert RoleMaster").Cause(err).Build()
	}

	// Ambil kembali data hasil Upsert secara penuh dari database
	upsertedEntity, err := s.queryRepo.FindByID(ctx, entity.Id)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve upserted RoleMaster").Cause(err).Build()
	}
	return mapEntityToResponse(upsertedEntity), nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build()
	}

	// Cek apakah record ada sebelum dihapus untuk idempotency dan pre-delete logic
	existing, err := s.queryRepo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalError().Message("Failed to retrieve RoleMaster before deletion").Cause(err).Build()
	}
	if existing == nil {
		return nil
	} // Idempotent: jika tidak ada, anggap berhasil

	if err := s.cmdRepo.Delete(ctx, id); err != nil {
		return errors.InternalError().Message("Failed to delete RoleMaster").Cause(err).Build()
	}

	// TODO: Tambahkan invalidasi cache di sini jika ada

	return nil
}

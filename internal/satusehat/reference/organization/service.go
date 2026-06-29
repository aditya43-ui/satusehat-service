package organization

import "context"

type Service interface {
	GetByID(ctx context.Context, id string) (map[string]interface{}, error)
	Search(ctx context.Context, params OrganizationSearchParams) (map[string]interface{}, error)
	Create(ctx context.Context, payload interface{}) (map[string]interface{}, error)
	Update(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error)
	Patch(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (map[string]interface{}, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Search(ctx context.Context, params OrganizationSearchParams) (map[string]interface{}, error) {
	return s.repo.Search(ctx, params)
}

func (s *service) Create(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
	return s.repo.Create(ctx, payload)
}
func (s *service) Update(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	return s.repo.Update(ctx, id, payload)
}
func (s *service) Patch(ctx context.Context, id string, payload interface{}) (map[string]interface{}, error) {
	return s.repo.Patch(ctx, id, payload)
}

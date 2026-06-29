package organization

// OrganizationSearchParams menampung parameter untuk pencarian Organisasi
type OrganizationSearchParams struct {
	Name       string `form:"name"`
	PartOf     string `form:"partof"`     // ID Organisasi Parent
	Identifier string `form:"identifier"` // Identifier Organisasi (contoh: http://sys-ids.kemkes.go.id/organization/10000004|R220001)
}

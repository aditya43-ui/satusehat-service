package location

// LocationSearchParams menampung kriteria pencarian Ruangan/Lokasi Satu Sehat
type LocationSearchParams struct {
	Name         string `form:"name"`
	Organization string `form:"organization"` // ID Faskes pembuat
	Identifier   string `form:"identifier"`   // Format {System}|{Value}
}

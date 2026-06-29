package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateError menerjemahkan error dari go-playground/validator menjadi pesan bahasa Indonesia.
func TranslateError(err error) string {
	var ve validator.ValidationErrors

	// Cek apakah error merupakan tipe ValidationErrors dari go-playground
	if errors.As(err, &ve) {
		var errMsgs []string
		for _, fe := range ve {
			// fe.Field() mengambil nama field struct (atau bisa pakai fe.Param() untuk parameter)
			switch fe.Tag() {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' wajib diisi", fe.Field()))
			case "oneof":
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' harus bernilai salah satu dari: %s", fe.Field(), fe.Param()))
			case "min":
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' minimal harus %s karakter", fe.Field(), fe.Param()))
			case "max":
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' maksimal harus %s karakter", fe.Field(), fe.Param()))
			case "email":
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' harus berupa format email yang valid", fe.Field()))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("Kolom '%s' tidak valid pada validasi '%s'", fe.Field(), fe.Tag()))
			}
		}
		return strings.Join(errMsgs, ", ")
	}

	// Kembalikan error asli jika bukan error validasi
	return err.Error()
}

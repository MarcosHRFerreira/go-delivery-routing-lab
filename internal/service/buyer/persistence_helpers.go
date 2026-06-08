package buyer

import (
	"errors"
	"strings"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

func normalizeBuyerInput(name string, document string, phone string, email string) (string, string, string, string) {
	return strings.TrimSpace(name), digitsOnly(document), digitsOnly(phone), strings.ToLower(strings.TrimSpace(email))
}

func digitsOnly(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func mapBuyerConflictError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}

	switch {
	case strings.Contains(mysqlErr.Message, "uq_buyers_document"):
		return apperror.Conflict("buyer document already exists")
	case strings.Contains(mysqlErr.Message, "uq_buyers_email"):
		return apperror.Conflict("buyer email already exists")
	case strings.Contains(mysqlErr.Message, "uq_buyers_phone"):
		return apperror.Conflict("buyer phone already exists")
	default:
		return apperror.Conflict("buyer already exists")
	}
}

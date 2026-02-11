package graph

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ✅ This regex is NOT vulnerable to ReDoS because there's no repetition operator. It does not contain any quantifiers, nested groups, or alternation. It's a single character class.
var validTitleRegex = regexp.MustCompile(`[\p{L}\p{M}\p{N}]`)

const MaxBatchMutationSize = 9000
const MaxFolderNameLen = 1000

type Resolver struct {
	DB *pgxpool.Pool
}

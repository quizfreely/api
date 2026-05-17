package rest

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RESTHandler struct {
	DB                 *pgxpool.Pool
	Storage            *s3.Client
	UsercontentBucket  *string
	UsercontentBaseURL *string
}

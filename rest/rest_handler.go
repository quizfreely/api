package rest

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RESTHandler struct {
	DB *pgxpool.Pool
	Storage *s3.Client
	StorageUsercontentBucket string
}

package storage

import (
	"context"
	"time"

	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/nikita89756/testEffectiveMobile/internal/storage/postgres"
	"github.com/nikita89756/testEffectiveMobile/pkg/logger"
)


type Storage interface{
GetPersonByID(context.Context,int)(*model.Person, error)
DeletePersonByID(context.Context, int) error
GetPersonsByFilter(context.Context,model.Person,int,int) ([]model.Person,error)
CreatePerson(context.Context, *model.Person) error
UpdatePersonByID(context.Context,*model.Person) error
}

func NewStorage(connectionString string,logger logger.Logger, timeout time.Duration	) Storage {
	return postgres.NewPostgres(connectionString,logger,timeout)
}

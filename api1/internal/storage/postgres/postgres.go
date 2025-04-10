package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	customerrors "github.com/nikita89756/testEffectiveMobile/internal/errors"
	"github.com/nikita89756/testEffectiveMobile/internal/model"
	logger "github.com/nikita89756/testEffectiveMobile/pkg/logger"
)

type Postgres struct {
	db      *sql.DB
	logger  logger.Logger
	timeout time.Duration
}

func NewPostgres(connectionString string, logger logger.Logger, timeout time.Duration) *Postgres {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		panic(err)
	}
	logger.Info("Подключение к базе данных завершено")
	return &Postgres{
		db:      db,
		logger:  logger,
		timeout: timeout,
	}
}

func (p *Postgres) Close() {
	if err := p.db.Close(); err != nil {
		p.logger.Error("Ошибка закрытия подключения к базе данных", p.logger.String("error", err.Error()))
	}
	p.logger.Info("Подключение к базе данных закрыто")
}

func (p *Postgres) CreatePerson(ctx context.Context,person *model.Person) error {
	query := `INSERT INTO people (name,surname,patronymic, age ,nationality,gender,created_at,updated_at) VALUES ($1, $2, $3, $4, $5, $6,$7,$8) RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		p.logger.Error("Ошибка начала транзакции", p.logger.String("error", err.Error()))
		return err
	}
	now := time.Now()
	person.CreatedAt = now
	person.UpdatedAt = now

	defer tx.Rollback()

	age := sql.NullInt64{Valid: person.Age != 0}
	if person.Age != 0 {
		age.Int64 = int64(person.Age)
	}
	gender := sql.NullString{Valid: person.Gender != ""}
	if person.Gender != "" {
		gender.String = person.Gender
	}
	nationality := sql.NullString{Valid: person.Nationality != ""}
	if person.Nationality != "" {
		nationality.String = person.Nationality
	}
	row:= tx.QueryRowContext(ctx,query,person.Name,person.Surname,person.Patronymic,age,nationality,gender,person.CreatedAt,person.UpdatedAt)
	if err = row.Scan(&person.ID); err != nil {
		p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
		return err
	}
	if err = tx.Commit(); err != nil {
		p.logger.Error("Ошибка коммита транзакции", p.logger.String("error", err.Error()))
		return err
	}
	id := strconv.Itoa(person.ID)
	p.logger.Info("Создана запись в таблице person", p.logger.String("id", id))
	return nil

}

func (p *Postgres) GetPersonByID(ctx context.Context,id int) (*model.Person, error) {
	var(
		age sql.NullInt64
		nationality sql.NullString
		gender sql.NullString
	)
	query := `SELECT name,surname,patronymic, age ,nationality,gender,created_at,updated_at FROM people WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	person := &model.Person{ID:id}
	row := p.db.QueryRowContext(ctx, query, id)
	err:= row.Scan(&person.Name,&person.Surname,&person.Patronymic,&age,&nationality,&gender,&person.CreatedAt,&person.UpdatedAt)
	if errors.Is(err,sql.ErrNoRows){
		return person , customerrors.ErrPersonNotFound
	}
	if err != nil {
		return person,err
	}
	if age.Valid {
		person.Age = age.Int64
	}
	if nationality.Valid{
		person.Nationality = nationality.String
	}
	if gender.Valid{
		person.Gender = gender.String
	}
	p.logger.Info("Получена запись из таблицы people", p.logger.String("id", strconv.Itoa(id)))
	return person,nil
}

func (p *Postgres) DeletePersonByID(ctx context.Context,id int) error {
	ctx , cancel:= context.WithTimeout(ctx, p.timeout)

	defer cancel()
	tx , err := p.db.BeginTx(ctx,nil)

	defer tx.Rollback()
	if err != nil {
		p.logger.Error("Ошибка начала транзакции", p.logger.String("error", err.Error()))
		return err
	}
	query := `DELETE FROM people WHERE id = $1`
	res, err := tx.ExecContext(ctx, query, id)
	rowsAffected, err := res.RowsAffected()
	if rowsAffected ==0{
		p.logger.Debug("Нечего удалять из базы")
		return customerrors.ErrNothingToDelete
	}
	if err != nil {
		p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
		return err
	}
	if err = tx.Commit(); err != nil {
		p.logger.Error("Ошибка коммита транзакции", p.logger.String("error", err.Error()))
		return err
	}
	return nil
}

func (p *Postgres) UpdatePersonByID(ctx context.Context,person *model.Person) error {
	query := `UPDATE people SET name = $1,surname = $2,patronymic = $3,age = $4,nationality = $5,gender = $6,updated_at = $7 WHERE id = $8`
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		p.logger.Error("Ошибка начала транзакции", p.logger.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()
	now := time.Now()
	person.UpdatedAt = now
	age := sql.NullInt64{Valid: person.Age != 0}
	if person.Age != 0 {
		age.Int64 = int64(person.Age)
	}
	gender := sql.NullString{Valid: person.Gender != ""}
	if person.Gender != "" {
		gender.String = person.Gender
	}
	nationality := sql.NullString{Valid: person.Nationality != ""}
	if person.Nationality != "" {
		nationality.String = person.Nationality
	}
	res, err := tx.ExecContext(ctx, query, person.Name,person.Surname,person.Patronymic,age,nationality,gender,person.UpdatedAt,person.ID)
	rowsAffected, err := res.RowsAffected()
	if rowsAffected ==0{
		p.logger.Debug("Нечего обновлять в базе")
		return customerrors.ErrNothingToUpdate
	}
	if err != nil {
		p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
		return err
	}
	if err = tx.Commit(); err != nil {
		p.logger.Error("Ошибка коммита транзакции", p.logger.String("error", err.Error()))
		return err
	}
	return nil
	}


func (p *Postgres) GetPersonsByFilter(ctx context.Context,person model.Person,offset,limit int) ([]model.Person,error){
	args := make([]interface{}, 0)
	args = appendArgs(args, person)
	query := "SELECT id, name,surname,patronymic, age ,nationality,gender,created_at,updated_at FROM people WHERE ($1 = '' or name = $1) and($2 = '' or surname = $2) and ($3 = '' or patronymic = $3) and ($4 = 0 or age = $4) and ($5 = '' or nationality = $5) and ($6 = '' or gender = $6) ORDER BY id "
	
	// Обработать когда нет LIMIT и OFFSET
	if limit == 0 {
		query += " LIMIT ALL OFFSET $7"
		args = append(args, offset)
	}else{
		query += " LIMIT $7 OFFSET $8"
		p.logger.Info("Limit and offset", p.logger.String("limit", strconv.Itoa(limit)), p.logger.String("offset", strconv.Itoa(offset)))
		args = append(args, limit, offset)
	}
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	rows , err:=p.db.QueryContext(ctx,query,args...)

	if err != nil{
		p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
		return nil ,err
	}
	defer rows.Close()
	
	persons := make([]model.Person, 0)
	for rows.Next() {
			person := model.Person{}
			var age sql.NullInt64
			var gender sql.NullString
			var nationality sql.NullString
			var patronymic sql.NullString

			if err := rows.Scan(
				&person.ID,
				&person.Name,
				&person.Surname,
				&patronymic,
				&age,
				&gender,
				&nationality,
				&person.CreatedAt,
				&person.UpdatedAt,
			); err != nil {
				p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
				return nil, fmt.Errorf("repository find with filters scan failed: %w", err)
			}

			if patronymic.Valid {
				person.Patronymic = patronymic.String
			}
			if age.Valid {
				person.Age = age.Int64
			}
			if gender.Valid {
				person.Gender = gender.String
			}
			if nationality.Valid {
				person.Nationality = nationality.String
			}
			persons = append(persons, person)
		}
	if err := rows.Err(); err != nil {
		p.logger.Error("Ошибка выполнения запроса", p.logger.String("error", err.Error()))
		return nil, fmt.Errorf("Ошибка при сканировании: %w", err)
	}
	p.logger.Info("Получены записи из таблицы person", p.logger.String("count", strconv.Itoa(len(persons))))
	return persons, nil
}

func appendArgs(args []interface{}, person model.Person) []interface{} {
	args = append(args, person.Name , person.Surname, person.Patronymic,person.Age,person.Nationality,person.Gender)
	return args
}
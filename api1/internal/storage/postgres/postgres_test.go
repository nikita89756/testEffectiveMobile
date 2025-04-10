package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	customerrors "github.com/nikita89756/testEffectiveMobile/internal/errors"
	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreatePerson(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewPostgres(db, zap.NewNop(), 1*time.Second)

	type args struct {
		ctx    context.Context
		person model.Person
	}

	type mockBehavior func(args args, id int)
	tests := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int 
		wantErr      bool
	}{
		{
			name: "Success",
			args: args{
				ctx: context.Background(),
				person: model.Person{
					Name:        "John",
					Surname:     "Doe",
					Patronymic:  "Smith", 
					Age:         30,
					Nationality: "USA",
					Gender:      "male",
				},
			},
			id: 1, 
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0}
				if args.person.Age != 0 {
					expectedAge.Int64 = int64(args.person.Age)
				}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != ""}
				if args.person.Nationality != "" {
					expectedNationality.String = args.person.Nationality
				}
				expectedGender := sql.NullString{Valid: args.person.Gender != ""}
				if args.person.Gender != "" {
					expectedGender.String = args.person.Gender
				}

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

				
				mock.ExpectQuery("INSERT INTO people").WithArgs(
						args.person.Name,         
						args.person.Surname,      
						args.person.Patronymic,   
						expectedAge,              
						expectedNationality,      
						expectedGender,           
						sqlmock.AnyArg(),         
						sqlmock.AnyArg(),         
					).WillReturnRows(rows)

				mock.ExpectCommit() 
			},
			wantErr: false,
		},

			{
				name: "DB Query Error",
				args: args{
					ctx: context.Background(),
					person: model.Person{Name: "Fail", Surname: "DB"},
				},
				id: 0, 
				mockBehavior: func(args args, id int) {
					mock.ExpectBegin()

					expectedAge := sql.NullInt64{Valid: args.person.Age != 0}
					if args.person.Age != 0 { expectedAge.Int64 = int64(args.person.Age) }
					expectedNationality := sql.NullString{Valid: args.person.Nationality != ""}
					if args.person.Nationality != "" { expectedNationality.String = args.person.Nationality }
					expectedGender := sql.NullString{Valid: args.person.Gender != ""}
					if args.person.Gender != "" { expectedGender.String = args.person.Gender }

					mock.ExpectQuery("INSERT INTO people").
						WithArgs(
							args.person.Name, args.person.Surname, args.person.Patronymic,
							expectedAge, expectedNationality, expectedGender,
							sqlmock.AnyArg(), sqlmock.AnyArg(),
						).WillReturnError(sql.ErrConnDone) 

					mock.ExpectRollback() 
				},
				wantErr: true, 
			},


			{
				name: "Commit Error",
				args: args{
					ctx: context.Background(),
					person: model.Person{Name: "Fail", Surname: "Commit"},
				},
				id: 2, 
				mockBehavior: func(args args, id int) {
					mock.ExpectBegin()

					
					expectedAge := sql.NullInt64{Valid: args.person.Age != 0}
					if args.person.Age != 0 { expectedAge.Int64 = int64(args.person.Age) }
					expectedNationality := sql.NullString{Valid: args.person.Nationality != ""}
					if args.person.Nationality != "" { expectedNationality.String = args.person.Nationality }
					expectedGender := sql.NullString{Valid: args.person.Gender != ""}
					if args.person.Gender != "" { expectedGender.String = args.person.Gender }

					rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

					mock.ExpectQuery("INSERT INTO people").
						WithArgs(
							args.person.Name, args.person.Surname, args.person.Patronymic,
							expectedAge, expectedNationality, expectedGender,
							sqlmock.AnyArg(), sqlmock.AnyArg(),
						).WillReturnRows(rows)

					mock.ExpectCommit().WillReturnError(sql.ErrTxDone) 

					
				},
				wantErr: true,
			},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.args, tt.id) 

			err := r.CreatePerson(tt.args.ctx, &tt.args.person)

			if tt.wantErr {
				assert.Error(t, err) 
			} else {
				assert.NoError(t, err)                   
				assert.Equal(t, tt.id, tt.args.person.ID) 
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "Не все ожидания sqlmock были выполнены")
		})
	}
}

func TestGetPersonByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Не удалось создать sqlmock")
	defer db.Close()

	r := NewPostgres(db, zap.NewNop(), 1*time.Second)
	testID := 1
	now := time.Now()

	expectedPerson := &model.Person{
		ID:          testID,
		Name:        "Jane",
		Surname:     "Doe",
		Patronymic:  "Alex",
		Age:         25,
		Nationality: "CAN",
		Gender:      "female",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	selectCols := []string{"name", "surname", "patronymic", "age", "nationality", "gender", "created_at", "updated_at"}

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name         string
		mockBehavior func(mock sqlmock.Sqlmock, args args)
		args         args
		want         *model.Person
		wantErr      error 
	}{
		{
			name: "Success",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {

				rows := sqlmock.NewRows(selectCols).
					AddRow(
						expectedPerson.Name,
						expectedPerson.Surname,

						sql.NullString{String: expectedPerson.Patronymic, Valid: expectedPerson.Patronymic != ""},
						sql.NullInt64{Int64: expectedPerson.Age, Valid: expectedPerson.Age != 0},
						sql.NullString{String: expectedPerson.Nationality, Valid: expectedPerson.Nationality != ""},
						sql.NullString{String: expectedPerson.Gender, Valid: expectedPerson.Gender != ""},
						expectedPerson.CreatedAt,
						expectedPerson.UpdatedAt,
					)

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT name,surname,patronymic, age ,nationality,gender,created_at,updated_at FROM people WHERE id = $1`)).WithArgs(args.id).WillReturnRows(rows)
			},
			want:    expectedPerson,
			wantErr: nil,
		},
		{
			name: "Not Found",
			args: args{
				ctx: context.Background(),
				id:  testID + 1, 
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT name,surname,patronymic, age ,nationality,gender,created_at,updated_at FROM people WHERE id = $1`)).
					WithArgs(args.id).
					WillReturnError(sql.ErrNoRows) 
			},

			want:    &model.Person{ID: testID + 1}, 
			wantErr: customerrors.ErrPersonNotFound,  
		},
		{
			name: "DB Error on Query",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT name,surname,patronymic, age ,nationality,gender,created_at,updated_at FROM people WHERE id = $1`)).
					WithArgs(args.id).
					WillReturnError(errors.New("db query error")) 
			},
			want:    &model.Person{ID: testID},
			wantErr: errors.New("db query error"), 
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mock, tt.args)

			got, err := r.GetPersonByID(tt.args.ctx, tt.args.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr) || err.Error() == tt.wantErr.Error(), "Ожидалась ошибка '%v', получена '%v'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "Не все ожидания sqlmock были выполнены")
		})
	}
}

func TestDeletePersonByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewPostgres(db, zap.NewNop(), 2*time.Second)

	type args struct {
		ctx context.Context
		id  int
	}

	type mockBehavior func(args args)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		args          args
		wantErr       bool
		expectedError error 
	}{
		{
			name: "Success",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin() 
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM people WHERE id = $1")).WithArgs(args.id).WillReturnResult(sqlmock.NewResult(0, 1)) 

				mock.ExpectCommit()
			},
			wantErr:       false,
			expectedError: nil,
		},
		{
			name: "Not Found - Nothing to Delete",
			args: args{
				ctx: context.Background(),
				id:  99, 
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM people WHERE id = $1")).WithArgs(args.id).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			wantErr:       true,
			expectedError: customerrors.ErrNothingToDelete,
		},
		{
			name: "DB Exec Error",
			args: args{
				ctx: context.Background(),
				id:  2,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM people WHERE id = $1")).WithArgs(args.id).WillReturnError(sql.ErrConnDone)

				mock.ExpectRollback()
			},
			wantErr:       true,
			expectedError: sql.ErrConnDone,
		},
		{
			name: "Commit Error",
			args: args{
				ctx: context.Background(),
				id:  3,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM people WHERE id = $1")).WithArgs(args.id).WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit().WillReturnError(sql.ErrTxDone)

			},
			wantErr:       true,
			expectedError: sql.ErrTxDone,
		},
		{
			name: "Begin Transaction Error",
			args: args{
				ctx: context.Background(),
				id:  4,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin().WillReturnError(errors.New("failed to begin tx"))
			},
			wantErr:       true,
			expectedError: errors.New("failed to begin tx"), 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.args)

			err := r.DeletePersonByID(tt.args.ctx, tt.args.id)
			if tt.wantErr {
				assert.Error(t, err) 
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "Не все ожидания sqlmock были выполнены")
		})
	}
}

func TestUpdatePersonByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewPostgres(db, zap.NewNop(), 2*time.Second)

	type args struct {
		ctx    context.Context
		person *model.Person
	}

	type mockBehavior func(args args)

	basePerson := &model.Person{
		ID:          1,
		Name:        "Jane",
		Surname:     "Doe",
		Patronymic:  "Anne",
		Age:         31,
		Nationality: "CAN",
		Gender:      "female",
	}

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		args          args
		wantErr       bool
		expectedError error
	}{
		{
			name: "Success",
			args: args{
				ctx:    context.Background(),
				person: basePerson,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0, Int64: int64(args.person.Age)}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != "", String: args.person.Nationality}
				expectedGender := sql.NullString{Valid: args.person.Gender != "", String: args.person.Gender}

				query := regexp.QuoteMeta("UPDATE people SET name = $1, surname = $2, patronymic = $3, age = $4, nationality = $5, gender = $6, updated_at = $7 WHERE id = $8")
				mock.ExpectExec(query).
					WithArgs(
						args.person.Name,
						args.person.Surname,
						args.person.Patronymic,
						expectedAge,
						expectedNationality,
						expectedGender,
						sqlmock.AnyArg(),
						args.person.ID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			wantErr:       false,
			expectedError: nil,
		},
		{
			name: "Not Found - Nothing to Update",
			args: args{
				ctx: context.Background(),
				person: &model.Person{
					ID:          99,
					Name:        "Nobody",
					Surname:     "Here",
					Age:         40,
					Nationality: "NA",
					Gender:      "other",
				},
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0, Int64: int64(args.person.Age)}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != "", String: args.person.Nationality}
				expectedGender := sql.NullString{Valid: args.person.Gender != "", String: args.person.Gender}

				query := regexp.QuoteMeta("UPDATE people SET name = $1, surname = $2, patronymic = $3, age = $4, nationality = $5, gender = $6, updated_at = $7 WHERE id = $8")
				mock.ExpectExec(query).
					WithArgs(
						args.person.Name, args.person.Surname, args.person.Patronymic,
						expectedAge, expectedNationality, expectedGender,
						sqlmock.AnyArg(),
						args.person.ID,
					).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectRollback()
			},
			wantErr:       true,
			expectedError: customerrors.ErrNothingToUpdate,
		},
		{
			name: "DB Exec Error",
			args: args{
				ctx:    context.Background(),
				person: basePerson,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0, Int64: int64(args.person.Age)}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != "", String: args.person.Nationality}
				expectedGender := sql.NullString{Valid: args.person.Gender != "", String: args.person.Gender}

				query := regexp.QuoteMeta("UPDATE people SET name = $1, surname = $2, patronymic = $3, age = $4, nationality = $5, gender = $6, updated_at = $7 WHERE id = $8")
				mock.ExpectExec(query).
					WithArgs(
						args.person.Name, args.person.Surname, args.person.Patronymic,
						expectedAge, expectedNationality, expectedGender,
						sqlmock.AnyArg(), args.person.ID,
					).
					WillReturnError(sql.ErrConnDone)

				mock.ExpectRollback()
			},
			wantErr:       true,
			expectedError: sql.ErrConnDone,
		},
		{
			name: "RowsAffected Error",
			args: args{
				ctx:    context.Background(),
				person: basePerson,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0, Int64: int64(args.person.Age)}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != "", String: args.person.Nationality}
				expectedGender := sql.NullString{Valid: args.person.Gender != "", String: args.person.Gender}

				query := regexp.QuoteMeta("UPDATE people SET name = $1, surname = $2, patronymic = $3, age = $4, nationality = $5, gender = $6, updated_at = $7 WHERE id = $8")

				mock.ExpectExec(query).
					WithArgs(
						args.person.Name, args.person.Surname, args.person.Patronymic,
						expectedAge, expectedNationality, expectedGender,
						sqlmock.AnyArg(), args.person.ID,
					).
					WillReturnError(errors.New("simulated error before RowsAffected"))

				mock.ExpectRollback()
			},
			wantErr:       true,
			expectedError: errors.New("simulated error before RowsAffected"),
		},
		{
			name: "Commit Error",
			args: args{
				ctx:    context.Background(),
				person: basePerson,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				expectedAge := sql.NullInt64{Valid: args.person.Age != 0, Int64: int64(args.person.Age)}
				expectedNationality := sql.NullString{Valid: args.person.Nationality != "", String: args.person.Nationality}
				expectedGender := sql.NullString{Valid: args.person.Gender != "", String: args.person.Gender}

				query := regexp.QuoteMeta("UPDATE people SET name = $1, surname = $2, patronymic = $3, age = $4, nationality = $5, gender = $6, updated_at = $7 WHERE id = $8")
				mock.ExpectExec(query).
					WithArgs(
						args.person.Name, args.person.Surname, args.person.Patronymic,
						expectedAge, expectedNationality, expectedGender,
						sqlmock.AnyArg(), args.person.ID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit().WillReturnError(sql.ErrTxDone)

			},
			wantErr:       true,
			expectedError: sql.ErrTxDone,
		},
		{
			name: "Begin Transaction Error",
			args: args{
				ctx:    context.Background(),
				person: basePerson,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin().WillReturnError(errors.New("failed to begin tx"))
			},
			wantErr:       true,
			expectedError: errors.New("failed to begin tx"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentPerson := new(model.Person)
			*currentPerson = *tt.args.person

			tt.mockBehavior(args{ctx: tt.args.ctx, person: currentPerson})

			err := r.UpdatePersonByID(tt.args.ctx, currentPerson)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "Не все ожидания sqlmock были выполнены")
		})
	}
}



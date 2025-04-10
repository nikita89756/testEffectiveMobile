package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikita89756/testEffectiveMobile/internal/handlers"
	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockStorage struct{}
type mockCache struct{}
type mockAddonService struct{}

func (m *mockStorage) CreatePerson(ctx context.Context, p *model.Person) error {
    p.ID = 1
    return nil
}
func (m *mockStorage) GetPersonByID(ctx context.Context, id int) (*model.Person, error) {
    return &model.Person{ID: id, Name: "Test", Surname: "User"}, nil
}
func (m *mockStorage) DeletePersonByID(ctx context.Context, id int) error {
    return nil
}
func (m *mockStorage) UpdatePersonByID(ctx context.Context, p *model.Person) error {
    return nil
}
func (m *mockStorage) GetPersonsByFilter(ctx context.Context, filter model.Person, offset, limit int) ([]model.Person, error) {
    return []model.Person{
        {ID: 1, Name: "John", Surname: "Doe"},
    }, nil
}

func (m *mockStorage) Migrate(migrationsDir string) error {
    return nil
}
func (m *mockCache) GetPerson(ctx context.Context, name string) (*model.PersonStats, error) {
    return &model.PersonStats{}, nil
}
func (m *mockCache) SetPersonWithTTL(ctx context.Context, name string, stats model.PersonStats) error {
    return nil
}

func (m *mockAddonService) Addon(p *model.Person) error {
    p.Age = 30
    p.Gender = "male"
    p.Nationality = "USA"
    return nil
}

func TestCreatePerson(t *testing.T) {
    gin.SetMode(gin.TestMode)

    handler := handlers.NewHandler(&mockStorage{}, zap.NewNop(), &mockAddonService{}, &mockCache{})

    router := gin.New()
    router.POST("/api/persons", handler.CreatePerson)

    payload := model.Person{Name: "John", Surname: "Doe"}
    body, _ := json.Marshal(payload)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/persons", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var res map[string]int
    _ = json.Unmarshal(w.Body.Bytes(), &res)
    assert.Equal(t, 1, res["id"])
}

func TestGetPersonByID(t *testing.T) {
    handler := handlers.NewHandler(&mockStorage{}, zap.NewNop(), &mockAddonService{}, &mockCache{})

    router := gin.New()
    router.GET("/api/persons/:id", handler.FindPersonByID)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/persons/1", nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var person model.Person
    _ = json.Unmarshal(w.Body.Bytes(), &person)
    assert.Equal(t, "Test", person.Name)
}

func TestGetPersons(t *testing.T) {
    handler := handlers.NewHandler(&mockStorage{}, zap.NewNop(), &mockAddonService{}, &mockCache{})
    router := gin.New()
    router.GET("/api/persons", handler.GetPersons)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/persons", nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var people []model.Person
    _ = json.Unmarshal(w.Body.Bytes(), &people)
    assert.Len(t, people, 1)
    assert.Equal(t, "John", people[0].Name)
}
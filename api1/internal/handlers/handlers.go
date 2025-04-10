package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	services "github.com/nikita89756/testEffectiveMobile/internal/apis"
	cache "github.com/nikita89756/testEffectiveMobile/internal/cache"
	customerrors "github.com/nikita89756/testEffectiveMobile/internal/errors"
	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/nikita89756/testEffectiveMobile/internal/storage"
	"github.com/nikita89756/testEffectiveMobile/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	storage   storage.Storage
	logger    logger.Logger
	addOnServ services.AddonService
	cache     cache.Cache
}

func NewHandler(storage storage.Storage, logger logger.Logger, addOnServ services.AddonService, cache cache.Cache) *Handler {
	return &Handler{
		storage:   storage,
		logger:    logger,
		addOnServ: addOnServ,
		cache:     cache,
	}
}

// @Summary Get person by ID
// @Tags persons
// @Description Get detailed information about a person by ID
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 200 {object} model.Person
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons/{id} [get]
func (h *Handler) FindPersonByID(ctx *gin.Context) {
	h.logger.Debug("FindPersonByID opened")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ivalid ID format"})
		return
	}
	person, err := h.storage.GetPersonByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, customerrors.ErrPersonNotFound) {
			h.logger.Warn("Person not found", zap.Int("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		h.logger.Error("Error while getting user from database", zap.String("error", err.Error()))
	}
	h.logger.Info("Успешно получены данные о человеке", zap.Int("id", id))
	person.ID = id
	ctx.JSON(http.StatusOK, person)
}

// @Summary Delete person by ID
// @Tags persons
// @Description Delete person from the database by ID
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 200
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons/{id} [delete]
func (h *Handler) DeletePersonByID(ctx *gin.Context) {
	h.logger.Debug("DeletePersonByID opened")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ivalid ID format"})
		return
	}

	err = h.storage.DeletePersonByID(ctx.Request.Context(), id)
	if errors.Is(err, customerrors.ErrNothingToDelete) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		return
	}
	if err != nil {
		h.logger.Error("Не удалось удалить запись", zap.Int("id", id), zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	ctx.Status(http.StatusOK)

}

// @Summary Update person by ID
// @Tags persons
// @Description Update person's information by ID
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param person body model.Person true "Updated person info"
// @Success 200
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons/{id} [put]
func (h *Handler) UpdatePersonByID(ctx *gin.Context) {
	h.logger.Debug("UpdatePersonByID opened")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ivalid ID format"})
		return
	}
	var person model.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		h.logger.Error("Неверный формат запроса", zap.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	person2, err := h.storage.GetPersonByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, customerrors.ErrPersonNotFound) {
			h.logger.Warn("Person not found", zap.Int("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		h.logger.Error("Ошибка получения данных о человеке", zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get person"})
		return
	}
	createPerson(person2, &person)
	err = h.storage.UpdatePersonByID(ctx.Request.Context(), person2)
	if err != nil {
		h.logger.Error("Не удалось обновить запись", zap.Int("id", id), zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}
	ctx.Status(http.StatusOK)
}

func createPerson(person *model.Person, person2 *model.Person) {
	if person2.Name != "" {
		person.Name = person2.Name
	}
	if person2.Surname != "" {
		person.Surname = person2.Surname
	}
	if person2.Patronymic != "" {
		person.Patronymic = person2.Patronymic
	}
	if person2.Age != 0 {
		person.Age = person2.Age
	}
	if person2.Nationality != "" {
		person.Nationality = person2.Nationality
	}
	if person2.Gender != "" {
		person.Gender = person2.Gender
	}
}

// @Summary Get persons with filters
// @Tags persons
// @Description Get list of persons with optional filters
// @Accept json
// @Produce json
// @Param name query string false "Name"
// @Param surname query string false "Surname"
// @Param patronymic query string false "Patronymic"
// @Param age query int false "Age"
// @Param gender query string false "Gender (male or female)"
// @Param nationality query string false "Nationality"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} model.Person
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons [get]
func (h *Handler) GetPersons(ctx *gin.Context) {
	age, err := strconv.Atoi(ctx.Query("age"))
	if err != nil {

	}
	person := model.Person{
		Name:        ctx.Query("name"),
		Surname:     ctx.Query("surname"),
		Patronymic:  ctx.Query("patronymic"),
		Age:         int64(age),
		Nationality: ctx.Query("nationality"),
		Gender:      ctx.Query("gender"),
	}
	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil {
		h.logger.Debug("Неверный формат лимита")
		limit = 0
	}
	offset, err := strconv.Atoi(ctx.Query("offset"))
	if err != nil {
		h.logger.Debug("Неверный формат смещения")
		offset = 0
	}
	if person.Gender != "male" && person.Gender != "female" && person.Gender != "" {
		h.logger.Debug("Неверный формат пола")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gender format"})
		return
	}
	persons, err := h.storage.GetPersonsByFilter(ctx.Request.Context(), person, offset, limit)
	if err != nil {
		h.logger.Error("Ошибка получения данных о людях", zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get persons"})
		return
	}
	if len(persons) == 0 {
		h.logger.Warn("Нет людей с такими данными")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No persons found"})
		return
	}
	h.logger.Info("Успешно получены данные о людях", zap.Int("count", len(persons)))
	ctx.JSON(http.StatusOK, persons)
}

// @Summary Create a new person
// @Tags persons
// @Description Add a new person to the database
// @Accept json
// @Produce json
// @Param person body model.Person true "Person info"
// @Success 200 {object} map[string]int
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons [post]
func (h *Handler) CreatePerson(ctx *gin.Context) {
	var person model.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		h.logger.Debug("Ошибка при парсинге JSON", zap.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	if person.Surname == "" || person.Name == "" {
		h.logger.Debug("Неверный формат имени или фамилии")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name or surname"})
		return
	}

	perstats, err := h.cache.GetPerson(ctx.Request.Context(), person.Name)
	if err != nil {
		h.logger.Error("Ошибка в кэше", zap.String("error", err.Error()))
		err = h.addOnServ.Addon(&person)
		if err != nil {
			h.logger.Error("Ошибка в аддоне", zap.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Addon error"})
			return
		}
		if person.Age != 0 && person.Gender != "" && person.Nationality != "" {
			err = h.cache.SetPersonWithTTL(ctx.Request.Context(), person.Name, model.PersonStats{
				Age:         person.Age,
				Gender:      person.Gender,
				Nationality: person.Nationality,
			})
			if err != nil {
				h.logger.Error("Ошибка в кэше", zap.String("error", err.Error()))
			}
		}
	} else {
		h.logger.Info("Успешно получен человек из кэша")
		person.Age = perstats.Age
		person.Gender = perstats.Gender
		person.Nationality = perstats.Nationality

	}

	err = h.storage.CreatePerson(ctx.Request.Context(), &person)
	if err != nil {
		h.logger.Error("Ошибка создания человека", zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person"})
		return
	}
	h.logger.Info("Успешно создан человек", zap.Int("id", person.ID))
	ctx.JSON(http.StatusOK, gin.H{"id": person.ID})
}

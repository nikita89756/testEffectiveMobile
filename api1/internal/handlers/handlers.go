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

// @Summary Получение данных о человеке по ID
// @Tags persons
// @Description Получение подробныйх данных о человеке по ID
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
			h.logger.Info("Person not found", zap.Int("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		h.logger.Error("Error while getting user from database", zap.String("error", err.Error()))
	}
	h.logger.Info("Успешно получены данные о человеке", zap.Int("id", id))
	person.ID = id
	ctx.JSON(http.StatusOK, person)
}

// @Summary Удаление человека
// @Tags persons
// @Description Удаление человека из базы данных по ID
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
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Ivalid ID format"})
		return
	}

	err = h.storage.DeletePersonByID(ctx.Request.Context(), id)
	if errors.Is(err, customerrors.ErrNothingToDelete) {
		ctx.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Person not found"})
		return
	}
	if err != nil {
		h.logger.Error("Не удалось удалить запись", zap.Int("id", id), zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Faild to delete"})
		return
	}
	ctx.Status(http.StatusOK)

}

// @Summary Обновление данных о человеке
// @Tags persons
// @Description Обновение данных о человеке по ID
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param person body model.PersonUpdateRequest true "Updated person info"
// @Success 200
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons/{id} [put]
func (h *Handler) UpdatePersonByID(ctx *gin.Context) {
	h.logger.Debug("UpdatePersonByID opened")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Ivalid ID format"})
		return
	}
	var person model.PersonUpdateRequest
	if err := ctx.ShouldBindJSON(&person); err != nil {
		h.logger.Error("Неверный формат запроса", zap.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid JSON"})
		return
	}
	person2, err := h.storage.GetPersonByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, customerrors.ErrPersonNotFound) {
			h.logger.Warn("Person not found", zap.Int("id", id))
			ctx.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Person not found"})
			return
		}
		h.logger.Error("Ошибка получения данных о человеке", zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError,model.ErrorResponse{Error: "Internal server error"})
		return
	}
	createPerson(person2, &person)
	err = h.storage.UpdatePersonByID(ctx.Request.Context(), person2)
	if err != nil {
		h.logger.Error("Не удалось обновить запись", zap.Int("id", id), zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Faild to update"})
		return
	}
	ctx.Status(http.StatusOK)
}

func createPerson(person *model.Person, person2 *model.PersonUpdateRequest) {
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

// @Summary Получить список человек
// @Tags persons
// @Description Возвращает список человек учитывая фильтры
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
		if limit < 0 {
			limit = 0
		}
		limit = 0
	}
	offset, err := strconv.Atoi(ctx.Query("offset"))
	if err != nil {
		h.logger.Debug("Неверный формат смещения")
		if offset < 0 {
			offset = 0
		}
		offset = 0
	}
	if person.Gender != "male" && person.Gender != "female" && person.Gender != "" {
		h.logger.Debug("Неверный формат пола")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid gender format"})
		return
	}
	persons, err := h.storage.GetPersonsByFilter(ctx.Request.Context(), person, offset, limit)
	if err != nil {
		h.logger.Error("Ошибка получения данных о людях", zap.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal server error"})
		return
	}
	if len(persons) == 0 {
		h.logger.Warn("Нет людей с такими данными")
		ctx.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Persons not found"})
		return
	}
	h.logger.Info("Успешно получены данные о людях", zap.Int("count", len(persons)))
	ctx.JSON(http.StatusOK, persons)
}

// @Summary Создает нового пользователя.
// @Tags persons
// @Description Добавление и обогащение данными ФИО.
// @Accept json
// @Produce json
// @Param person body model.PersonCreateRequest true "Person info"
// @Success 200 {object} model.IdResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /persons [post]
func (h *Handler) CreatePerson(ctx *gin.Context) {
	var persReq model.PersonCreateRequest
	var person model.Person

	if err := ctx.ShouldBindJSON(&persReq); err != nil {
		h.logger.Debug("Ошибка при парсинге JSON", zap.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid JSON"})
		return
	}
	if persReq.Surname == "" || persReq.Name == "" {
		h.logger.Debug("Неверный формат имени или фамилии")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid name or surname format"})
		return
	}
	person.Name = persReq.Name
	person.Surname = persReq.Surname
	person.Patronymic = persReq.Patronymic

	perstats, err := h.cache.GetPerson(ctx.Request.Context(), person.Name)
	if err != nil {
		h.logger.Error("Ошибка в кэше", zap.String("error", err.Error()))
		err = h.addOnServ.Addon(&person)
		if err != nil {
			h.logger.Error("Ошибка в аддоне", zap.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal server error"})
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
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Faild to create person"})
		return
	}
	h.logger.Info("Успешно создан человек", zap.Int("id", person.ID))
	ctx.JSON(http.StatusOK, model.IdResponse{ID: person.ID})
}

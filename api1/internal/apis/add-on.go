package services

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/nikita89756/testEffectiveMobile/pkg/logger"
	"go.uber.org/zap"
)

const (
	agifyURL       = "https://api.agify.io/?name="
	genderizeURL   = "https://api.genderize.io/?name="
	nationalizeURL = "https://api.nationalize.io/?name="
)

type AddonService interface {
	Addon(*model.Person) error
}

type Addon struct {
	client *http.Client
	logger logger.Logger
}

func NewAddonService(apiTimeout time.Duration,logger logger.Logger) AddonService {
	client := &http.Client{
		Timeout: apiTimeout,
	}


	return &Addon{
		client: client,
		logger: logger,
	}
}

func (s *Addon) Addon(person *model.Person) error {
	var wg sync.WaitGroup
	wg.Add(3)
	go s.getAgify(&wg, person, person.Name)
	go s.getGenderize(&wg, person, person.Name)
	go s.getNationalize(&wg, person, person.Name)
	wg.Wait()
	s.logger.Info("Добавлены дополнительные поля",zap.String("name", person.Name),zap.Int("age", int(person.Age)),zap.String("gender", person.Gender),zap.String("nationality", person.Nationality))
	return nil
}

func (s *Addon) getAgify(wg *sync.WaitGroup, person *model.Person, name string) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), s.client.Timeout)

		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, agifyURL+name, nil)

		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Error("Ошибка в getAgify", zap.String("name", name), zap.String("error", err.Error()))
			person.Age = 0
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("Ошибка в getAgify", zap.String("name", name), zap.Int("status_code", resp.StatusCode))
			person.Age = 0
			return
		}
		var agifyResponse model.Age

		if err := json.NewDecoder(resp.Body).Decode(&agifyResponse); err != nil {
			s.logger.Error("Ошибка в getAgify  при декодировании", zap.String("name", name), zap.String("error", err.Error()))
			person.Age = 0
			return
		}
		s.logger.Info("Успешный ответ от getAgify", zap.String("name", name), zap.Int("age", agifyResponse.Age))
		person.Age = int64(agifyResponse.Age)

}

func (s *Addon) getGenderize(wg *sync.WaitGroup, person *model.Person, name string) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), s.client.Timeout)

		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, genderizeURL+name, nil)

		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Error("Ошибка в getGenderize", zap.String("name", name), zap.String("error", err.Error()))
			person.Gender = ""
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("Ошибка в getGenderize", zap.String("name", name), zap.Int("status_code", resp.StatusCode))
			person.Gender = ""
			return
		}
		var genderizeResponse model.Gender
		if err := json.NewDecoder(resp.Body).Decode(&genderizeResponse); err != nil {
			s.logger.Error("Ошибка в getGenderize  при декодировании", zap.String("name", name), zap.String("error", err.Error()))
			person.Gender = ""
			return
		}
		s.logger.Info("Успешный ответ от getGenderize", zap.String("name", name), zap.String("gender", genderizeResponse.Gender))
		person.Gender = genderizeResponse.Gender
}

func (s *Addon) getNationalize(wg *sync.WaitGroup, person *model.Person, name string) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), s.client.Timeout)

		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nationalizeURL+name, nil)

		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Error("Ошибка в getNationalize", zap.String("name", name), zap.String("error", err.Error()))
			person.Nationality = ""
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("Ошибка в getNationalize", zap.String("name", name), zap.Int("status_code", resp.StatusCode))
			person.Nationality = ""
			return
		}
		var nationalizeResponse model.CountryList
		if err := json.NewDecoder(resp.Body).Decode(&nationalizeResponse); err != nil {
			s.logger.Error("Ошибка в getNationalize  при декодировании", zap.String("name", name), zap.String("error", err.Error()))
			person.Nationality = ""
			return
		}
		probability := nationalizeResponse.Countries[0].Probability
		countryId := nationalizeResponse.Countries[0].CountryID

		for i := 1; i < len(nationalizeResponse.Countries); i++ {
			if nationalizeResponse.Countries[i].Probability > probability {
				probability = nationalizeResponse.Countries[i].Probability
				countryId = nationalizeResponse.Countries[i].CountryID
			}
		}
		person.Nationality = countryId

		s.logger.Info("Успешный ответ от getNationalize", zap.String("name", name), zap.String("country", countryId))
}

package service

import (
	"context"

	"nla/internal/model"
	mongoRepo "nla/internal/mongo"
)

type RatingService struct {
	repo *mongoRepo.RatingRepo
}

func NewRatingService(repo *mongoRepo.RatingRepo) *RatingService {
	return &RatingService{repo: repo}
}

func (s *RatingService) GetByIssuer(ctx context.Context, issuer string) (*model.IssuerRatingResponse, error) {
	ratings, err := s.repo.GetByIssuer(ctx, issuer)
	if err != nil {
		return nil, err
	}

	resp := &model.IssuerRatingResponse{
		Issuer:  issuer,
		Ratings: ratings,
	}

	if len(ratings) > 0 {
		maxScore := 0
		for _, r := range ratings {
			if r.Score > maxScore {
				maxScore = r.Score
			}
		}
		resp.Score = maxScore
	}

	return resp, nil
}

func (s *RatingService) GetAll(ctx context.Context) (map[string]*model.IssuerRatingResponse, error) {
	ratings, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*model.IssuerRatingResponse)
	for _, r := range ratings {
		resp, ok := result[r.Issuer]
		if !ok {
			resp = &model.IssuerRatingResponse{
				Issuer:  r.Issuer,
				Ratings: []model.IssuerRating{},
			}
			result[r.Issuer] = resp
		}
		resp.Ratings = append(resp.Ratings, r)
		if r.Score > resp.Score {
			resp.Score = r.Score
		}
	}

	return result, nil
}

func (s *RatingService) Upsert(ctx context.Context, rating *model.IssuerRating) error {
	return s.repo.Upsert(ctx, rating)
}

func (s *RatingService) BulkUpsert(ctx context.Context, ratings []model.IssuerRating) error {
	return s.repo.BulkUpsert(ctx, ratings)
}

func (s *RatingService) Delete(ctx context.Context, issuer, agency string) error {
	return s.repo.Delete(ctx, issuer, agency)
}

// SeedDefaults populates default credit ratings for well-known Russian issuers.
// Only inserts if the collection is empty.
func (s *RatingService) SeedDefaults(ctx context.Context) error {
	existing, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	defaults := []model.IssuerRating{
		// ОФЗ (государственные)
		{Issuer: "Минфин России", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Минфин России", Agency: "Эксперт РА", Rating: "ruAAA", Score: 10},

		// Системообразующие банки
		{Issuer: "Сбербанк", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "ВТБ", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Газпромбанк", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Альфа-Банк", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Россельхозбанк", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Тинькофф Банк", Agency: "Эксперт РА", Rating: "ruAA", Score: 8},
		{Issuer: "Совкомбанк", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "МКБ", Agency: "АКРА", Rating: "AA-(RU)", Score: 7},
		{Issuer: "Банк ДОМ.РФ", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Промсвязьбанк", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},

		// Крупные корпорации
		{Issuer: "Газпром", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Роснефть", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "РЖД", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Транснефть", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Ростелеком", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "МТС", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Магнит", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "Лента", Agency: "АКРА", Rating: "A+(RU)", Score: 6},
		{Issuer: "ВК", Agency: "Эксперт РА", Rating: "ruA+", Score: 6},
		{Issuer: "Яндекс", Agency: "Эксперт РА", Rating: "ruAA-", Score: 7},
		{Issuer: "X5 Group", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "ЛУКОЙЛ", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Норникель", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "НЛМК", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Северсталь", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "ЕВРАЗ", Agency: "Эксперт РА", Rating: "ruAA", Score: 8},
		{Issuer: "Полюс", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "АЛРОСА", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Русал", Agency: "АКРА", Rating: "AA-(RU)", Score: 7},
		{Issuer: "ФосАгро", Agency: "АКРА", Rating: "AA(RU)", Score: 8},

		// Девелоперы
		{Issuer: "ПИК", Agency: "АКРА", Rating: "A+(RU)", Score: 6},
		{Issuer: "Самолет", Agency: "АКРА", Rating: "A(RU)", Score: 5},
		{Issuer: "ЛСР", Agency: "Эксперт РА", Rating: "ruA+", Score: 6},
		{Issuer: "Эталон", Agency: "Эксперт РА", Rating: "ruA", Score: 5},

		// Лизинг и финансы
		{Issuer: "Система", Agency: "АКРА", Rating: "AA-(RU)", Score: 7},
		{Issuer: "ДОМ.РФ", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "ВЭБ.РФ", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "ГТЛК", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "Европлан", Agency: "Эксперт РА", Rating: "ruA+", Score: 6},
		{Issuer: "Балтийский лизинг", Agency: "Эксперт РА", Rating: "ruA+", Score: 6},

		// Субфедеральные
		{Issuer: "Москва", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Московская обл", Agency: "АКРА", Rating: "AA+(RU)", Score: 9},
		{Issuer: "Санкт-Петербург", Agency: "АКРА", Rating: "AAA(RU)", Score: 10},
		{Issuer: "Свердловская обл", Agency: "АКРА", Rating: "AA(RU)", Score: 8},
		{Issuer: "Краснодарский край", Agency: "АКРА", Rating: "AA-(RU)", Score: 7},

		// Средние компании
		{Issuer: "РЕСО-Лизинг", Agency: "Эксперт РА", Rating: "ruA+", Score: 6},
		{Issuer: "Сэтл Групп", Agency: "Эксперт РА", Rating: "ruA", Score: 5},
		{Issuer: "Whoosh", Agency: "Эксперт РА", Rating: "ruBBB+", Score: 4},
		{Issuer: "Позитив", Agency: "Эксперт РА", Rating: "ruA", Score: 5},
		{Issuer: "Селектел", Agency: "АКРА", Rating: "A-(RU)", Score: 5},
		{Issuer: "Сегежа", Agency: "АКРА", Rating: "BB+(RU)", Score: 3},
		{Issuer: "МВидео", Agency: "АКРА", Rating: "A-(RU)", Score: 5},
		{Issuer: "АФК Система", Agency: "АКРА", Rating: "AA-(RU)", Score: 7},

		// ВДО (высокодоходные)
		{Issuer: "Калита", Agency: "Эксперт РА", Rating: "ruBBB-", Score: 3},
		{Issuer: "КарМани", Agency: "Эксперт РА", Rating: "ruBBB-", Score: 3},
		{Issuer: "МФК Займер", Agency: "Эксперт РА", Rating: "ruBBB", Score: 4},
	}

	return s.repo.BulkUpsert(ctx, defaults)
}

package repository

import "fmt"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

type Drug struct {
	DrugID                 int
	DrugName               string
	RecommendedAdultDoseMg float64
	MaxDailyDoseMg         float64
	ShortInfo              string
	DrugStatus             string
	ImageKey               string
	VideoKey               string
	LikedByUserIDs         []int
}

type Repository struct {
	drugs []Drug
}

func NewRepository() (*Repository, error) {
	drugs := []Drug{
		{
			DrugID:                 1,
			DrugName:               "Парацетамол",
			RecommendedAdultDoseMg: 500,
			MaxDailyDoseMg:         4000,
			ShortInfo:              "Ненаркотический анальгетик и антипиретик первой линии. Ингибирует циклооксигеназу преимущественно в центральной нервной системе. Разовая доза для взрослого 500 мг, интервал между приёмами не менее четырёх часов. Детская доза пересчитывается по площади поверхности тела.",
			DrugStatus:             StatusPublished,
			ImageKey:               "paracetamol.jpg",
			VideoKey:               "paracetamol.mp4",
			LikedByUserIDs:         []int{1, 2, 4, 7},
		},
		{
			DrugID:                 2,
			DrugName:               "Ибупрофен",
			RecommendedAdultDoseMg: 400,
			MaxDailyDoseMg:         1200,
			ShortInfo:              "Нестероидный противовоспалительный препарат группы пропионовой кислоты. Обладает противовоспалительным, анальгезирующим и жаропонижающим действием. Разовая доза для взрослого 400 мг, кратность приёма до трёх раз в сутки, приём после еды.",
			DrugStatus:             StatusPublished,
			ImageKey:               "ibuprofen.jpg",
			VideoKey:               "ibuprofen.mp4",
			LikedByUserIDs:         []int{2, 3},
		},
		{
			DrugID:                 3,
			DrugName:               "Амоксициллин",
			RecommendedAdultDoseMg: 500,
			MaxDailyDoseMg:         1500,
			ShortInfo:              "Полусинтетический пенициллин широкого спектра действия. Бактерицидный антибиотик, нарушающий синтез клеточной стенки. Стандартная схема для взрослого составляет 500 мг три раза в сутки курсом от пяти до семи дней.",
			DrugStatus:             StatusPublished,
			ImageKey:               "amoxicillin.jpg",
			VideoKey:               "amoxicillin.mp4",
			LikedByUserIDs:         []int{1, 5, 6},
		},
		{
			DrugID:                 4,
			DrugName:               "Цефтриаксон",
			RecommendedAdultDoseMg: 1000,
			MaxDailyDoseMg:         4000,
			ShortInfo:              "Цефалоспорин третьего поколения для парентерального применения. Применяется при тяжёлых бактериальных инфекциях. Взрослая доза 1000 мг один раз в сутки, при тяжёлом течении суточная доза увеличивается до 4000 мг в два введения.",
			DrugStatus:             StatusPublished,
			ImageKey:               "ceftriaxone.jpg",
			VideoKey:               "ceftriaxone.mp4",
			LikedByUserIDs:         []int{3},
		},
		{
			DrugID:                 5,
			DrugName:               "Метотрексат",
			RecommendedAdultDoseMg: 20,
			MaxDailyDoseMg:         25,
			ShortInfo:              "Антиметаболит, дозируемый строго по площади поверхности тела. Цитостатик группы антагонистов фолиевой кислоты. Расчёт дозы по площади поверхности тела обязателен, отклонение от расчётной дозы недопустимо. Для взрослого ориентировочная разовая доза составляет 20 мг.",
			DrugStatus:             StatusPublished,
			ImageKey:               "methotrexate.jpg",
			VideoKey:               "methotrexate.mp4",
			LikedByUserIDs:         []int{2, 4, 5, 6, 8},
		},
		{
			DrugID:                 6,
			DrugName:               "Азитромицин",
			RecommendedAdultDoseMg: 500,
			MaxDailyDoseMg:         500,
			ShortInfo:              "Создаёт высокие концентрации в тканях и сохраняет активность до пяти суток после отмены. Один раз в сутки в течение трёх дней.",
			DrugStatus:             StatusDraft,
			ImageKey:               "azithromycin.jpg",
			VideoKey:               "azithromycin.mp4",
			LikedByUserIDs:         []int{},
		},
		{
			DrugID:                 7,
			DrugName:               "Дексаметазон",
			RecommendedAdultDoseMg: 8,
			MaxDailyDoseMg:         20,
			ShortInfo:              "Синтетический глюкокортикостероид длительного действия. Оказывает противовоспалительное и противоаллергическое действие. Взрослая разовая доза 8 мг, суточная доза не более 20 мг.",
			DrugStatus:             StatusDeleted,
			ImageKey:               "dexamethasone.jpg",
			VideoKey:               "dexamethasone.mp4",
			LikedByUserIDs:         []int{7},
		},
		{
			DrugID:                 8,
			DrugName:               "Ципрофлоксацин",
			RecommendedAdultDoseMg: 250,
			MaxDailyDoseMg:         1000,
			ShortInfo:              "Фторхинолон второго поколения широкого спектра действия. Подавляет ДНК-гиразу бактерий и действует бактерицидно. Взрослая доза 250 мг два раза в сутки, при тяжёлых инфекциях разовая доза увеличивается до 750 мг.",
			DrugStatus:             StatusPublished,
			ImageKey:               "ciprofloxacin.jpg",
			VideoKey:               "ciprofloxacin.mp4",
			LikedByUserIDs:         []int{1, 3, 8},
		},
		{
			DrugID:                 9,
			DrugName:               "Омепразол",
			RecommendedAdultDoseMg: 20,
			MaxDailyDoseMg:         40,
			ShortInfo:              "Ингибитор протонного насоса, снижающий секрецию соляной кислоты в желудке. Капсула содержит кишечнорастворимые гранулы, поэтому её нельзя разжёвывать. Взрослая доза 20 мг один раз в сутки утром до еды.",
			DrugStatus:             StatusPublished,
			ImageKey:               "omeprazole.jpg",
			VideoKey:               "omeprazole.mp4",
			LikedByUserIDs:         []int{2, 5, 6, 9},
		},
		{
			DrugID:                 10,
			DrugName:               "Диклофенак",
			RecommendedAdultDoseMg: 50,
			MaxDailyDoseMg:         150,
			ShortInfo:              "Нестероидный противовоспалительный препарат из производных фенилуксусной кислоты. Обладает выраженным обезболивающим и противовоспалительным действием при болях в суставах и мышцах. Взрослая разовая доза 50 мг, суточная не более 150 мг.",
			DrugStatus:             StatusPublished,
			ImageKey:               "diclofenac.jpg",
			VideoKey:               "diclofenac.mp4",
			LikedByUserIDs:         []int{4, 7},
		},
	}

	return &Repository{drugs: drugs}, nil
}

func (r *Repository) GetPublishedDrugs() ([]Drug, error) {
	result := make([]Drug, 0, len(r.drugs))

	for _, drug := range r.drugs {
		if drug.DrugStatus == StatusPublished {
			result = append(result, drug)
		}
	}

	return result, nil
}

func (r *Repository) GetPublishedDrugsByAdultDose(minAdultDoseMg, maxAdultDoseMg float64) ([]Drug, error) {
	published, err := r.GetPublishedDrugs()
	if err != nil {
		return []Drug{}, err
	}

	result := make([]Drug, 0, len(published))

	for _, drug := range published {
		if drug.RecommendedAdultDoseMg >= minAdultDoseMg && drug.RecommendedAdultDoseMg <= maxAdultDoseMg {
			result = append(result, drug)
		}
	}

	return result, nil
}

func (r *Repository) GetDrugByID(drugID int) (Drug, error) {
	for _, drug := range r.drugs {
		if drug.DrugID == drugID && drug.DrugStatus == StatusPublished {
			return drug, nil
		}
	}

	return Drug{}, fmt.Errorf("препарат с идентификатором %d не найден или не опубликован", drugID)
}

func (r *Repository) GetNextDrug(drugID int) (Drug, error) {
	published, err := r.GetPublishedDrugs()
	if err != nil {
		return Drug{}, err
	}

	if len(published) == 0 {
		return Drug{}, fmt.Errorf("опубликованные препараты отсутствуют")
	}

	for index, drug := range published {
		if drug.DrugID == drugID {
			return published[(index+1)%len(published)], nil
		}
	}

	return published[0], nil
}

func (r *Repository) GetFirstPublishedDrug() (Drug, error) {
	published, err := r.GetPublishedDrugs()
	if err != nil {
		return Drug{}, err
	}

	if len(published) == 0 {
		return Drug{}, fmt.Errorf("опубликованные препараты отсутствуют")
	}

	return published[0], nil
}

func (r *Repository) GetDraftDrug() (Drug, error) {
	for _, drug := range r.drugs {
		if drug.DrugStatus == StatusDraft {
			return drug, nil
		}
	}

	return Drug{}, fmt.Errorf("препарат в статусе черновик отсутствует")
}

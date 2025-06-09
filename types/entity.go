package types

import "math"

// INF — константа для обозначения бесконечности в контексте планирования задач.
const INF = math.MaxInt32

// Job — описание задачи для планировщика.
type Job struct {
	Type     int
	Instance int
	Setup    int
	Process  int
	Deadline int
}

// GenerateRequest — запрос на генерацию задачи со всеми параметрами.
type GenerateRequest struct {
	NumTypes      int `json:"numTypes"`      // Число типов (T)
	PiecesPerType int `json:"piecesPerType"` // Копий каждого типа (P)
	MinSetup      int `json:"minSetup"`      // Минимальное время переналадки
	MaxSetup      int `json:"maxSetup"`      // Максимальное время переналадки
	MinProcess    int `json:"minProcess"`    // Минимальное время обработки
	MaxProcess    int `json:"maxProcess"`    // Максимальное время обработки
}

// Step - cтруктура для хранения шага планировщика для дальнейшего построения графика.
type Step struct {
	Type     int
	Instance int
	Start    int // момент начала setup
	SetupLen int
	ProcLen  int
}

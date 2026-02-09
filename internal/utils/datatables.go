// Package utils содержит вспомогательные функции для работы с DataTables.
// DataTables - это jQuery плагин для создания интерактивных таблиц с серверной обработкой.
package utils

import (
	"ctweb/internal/repositories"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// DataTablesRequest представляет параметры запроса от DataTables.
//
// DataTables отправляет POST запросы с параметрами:
//   - draw: номер запроса (для синхронизации)
//   - start: начальная позиция (offset)
//   - length: количество записей (limit)
//   - search[value]: глобальный поисковый запрос
//   - order[0][column]: индекс колонки для сортировки
//   - order[0][dir]: направление сортировки ("asc" или "desc")
//   - columns[i][data]: имя поля колонки
//   - columns[i][searchable]: можно ли искать по этой колонке
//   - columns[i][search][value]: значение для поиска по колонке
//
// Пример использования:
//
//	req := utils.ParseDataTablesRequest(c)
//	response := repo.FindAllWithPagination(req)
type DataTablesRequest struct {
	Draw   int    // Номер запроса (для синхронизации с клиентом)
	Start  int    // Начальная позиция (offset)
	Length int    // Количество записей (limit)
	Search string // Поисковый запрос (для фильтрации)
	Order  []struct {
		Column int    // Индекс колонки для сортировки
		Dir    string // Направление сортировки ("asc" или "desc")
	}
	Columns []struct {
		Data       string // Имя поля
		Searchable bool   // Можно ли искать по этому полю
		Search     struct {
			Value string // Значение для поиска
		}
	}
}

// ParseDataTablesRequest парсит параметры DataTables запроса из Gin контекста.
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - *DataTablesRequest: распарсенные параметры запроса
//
// Пример использования:
//
//	req := utils.ParseDataTablesRequest(c)
//	// req.Start, req.Length, req.Search и т.д.
func ParseDataTablesRequest(c *gin.Context) *DataTablesRequest {
	req := &DataTablesRequest{}

	// Парсим draw (номер запроса)
	if drawStr := c.PostForm("draw"); drawStr != "" {
		if draw, err := strconv.Atoi(drawStr); err == nil {
			req.Draw = draw
		}
	}

	// Парсим start (начальная позиция)
	if startStr := c.PostForm("start"); startStr != "" {
		if start, err := strconv.Atoi(startStr); err == nil {
			req.Start = start
		}
	}

	// Парсим length (количество записей)
	if lengthStr := c.PostForm("length"); lengthStr != "" {
		if length, err := strconv.Atoi(lengthStr); err == nil {
			req.Length = length
		}
	}

	// Парсим search[value] (глобальный поиск)
	req.Search = c.PostForm("search[value]")

	// Парсим order (сортировка)
	// DataTables может отправлять несколько order параметров: order[0][column], order[0][dir], order[1][column] и т.д.
	req.Order = []struct {
		Column int
		Dir    string
	}{}

	// Парсим все order параметры (начинаем с 0 и идём до тех пор, пока есть данные)
	for i := 0; i < 10; i++ { // Максимум 10 уровней сортировки (защита от бесконечного цикла)
		orderColStr := c.PostForm("order[" + strconv.Itoa(i) + "][column]")
		if orderColStr == "" {
			// Больше order параметров нет
			break
		}

		if orderCol, err := strconv.Atoi(orderColStr); err == nil {
			orderDir := c.PostForm("order[" + strconv.Itoa(i) + "][dir]")
			if orderDir == "" {
				orderDir = "asc" // По умолчанию ascending
			}
			req.Order = append(req.Order, struct {
				Column int
				Dir    string
			}{
				Column: orderCol,
				Dir:    orderDir,
			})
		}
	}

	// Парсим columns (колонки таблицы)
	// DataTables отправляет columns[0][data], columns[0][searchable], columns[0][search][value] и т.д.
	req.Columns = []struct {
		Data       string
		Searchable bool
		Search     struct {
			Value string
		}
	}{}

	// Парсим колонки
	// DataTables отправляет columns[0][data], columns[0][searchable], columns[0][search][value] и т.д.
	// Важно: первая колонка (checkbox) может иметь пустой data, но это не означает конец списка
	// Проверяем наличие параметров для каждой колонки
	for i := 0; i < 100; i++ { // Максимум 100 колонок (защита от бесконечного цикла)
		colData := c.PostForm("columns[" + strconv.Itoa(i) + "][data]")
		searchableStr := c.PostForm("columns[" + strconv.Itoa(i) + "][searchable]")

		// Если нет ни data, ни searchable, значит колонок больше нет
		// Но data может быть пустым для checkbox колонки, поэтому проверяем searchable
		// Если и searchable пустой, проверяем следующую колонку
		if colData == "" && searchableStr == "" {
			// Проверяем, есть ли следующая колонка (i+1)
			// Если есть, значит текущая колонка просто с пустым data (checkbox)
			nextColData := c.PostForm("columns[" + strconv.Itoa(i+1) + "][data]")
			nextSearchableStr := c.PostForm("columns[" + strconv.Itoa(i+1) + "][searchable]")
			if nextColData == "" && nextSearchableStr == "" {
				// Больше колонок нет
				break
			}
			// Если следующая колонка есть, продолжаем парсить текущую (даже с пустым data)
		}

		// Парсим searchable (может быть "true" или "false")
		searchable := searchableStr == "true"

		// Парсим search[value] для этой колонки
		searchValue := c.PostForm("columns[" + strconv.Itoa(i) + "][search][value]")

		req.Columns = append(req.Columns, struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}{
			Data:       colData, // Может быть пустым для checkbox колонки
			Searchable: searchable,
			Search: struct {
				Value string
			}{
				Value: searchValue,
			},
		})
	}

	// Логирование для отладки (временно, можно убрать после проверки)
	if len(req.Columns) == 0 {
		// Если колонок нет, логируем все параметры columns
		for i := 0; i < 5; i++ {
			colData := c.PostForm("columns[" + strconv.Itoa(i) + "][data]")
			searchableStr := c.PostForm("columns[" + strconv.Itoa(i) + "][searchable]")
			if colData != "" || searchableStr != "" {
				// Есть параметры, но они не были распарсены - это ошибка
			}
		}
	}

	return req
}

// ConvertToUserRepositoryRequest конвертирует DataTablesRequest в формат UserRepository.
//
// Параметры:
//   - req: запрос из ParseDataTablesRequest
//
// Возвращает:
//   - *repositories.UserDataTablesRequest: запрос для UserRepository
func ConvertToUserRepositoryRequest(req *DataTablesRequest) *repositories.UserDataTablesRequest {
	userReq := &repositories.UserDataTablesRequest{
		Start:  req.Start,
		Length: req.Length,
		Search: req.Search,
		Order: make([]struct {
			Column int
			Dir    string
		}, len(req.Order)),
		Columns: make([]struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}, len(req.Columns)),
	}

	// Копируем Order
	for i, order := range req.Order {
		userReq.Order[i] = struct {
			Column int
			Dir    string
		}{
			Column: order.Column,
			Dir:    order.Dir,
		}
	}

	// Копируем Columns
	for i, col := range req.Columns {
		userReq.Columns[i] = struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}{
			Data:       col.Data,
			Searchable: col.Searchable,
			Search: struct {
				Value string
			}{
				Value: col.Search.Value,
			},
		}
	}

	return userReq
}

// ConvertToGroupRepositoryRequest конвертирует DataTablesRequest в формат GroupRepository.
//
// Параметры:
//   - req: запрос из ParseDataTablesRequest
//
// Возвращает:
//   - *repositories.DataTablesRequest: запрос для GroupRepository
func ConvertToGroupRepositoryRequest(req *DataTablesRequest) *repositories.DataTablesRequest {
	groupReq := &repositories.DataTablesRequest{
		Start:  req.Start,
		Length: req.Length,
		Search: req.Search,
		Order: make([]struct {
			Column int
			Dir    string
		}, len(req.Order)),
		Columns: make([]struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}, len(req.Columns)),
	}

	// Копируем Order
	for i, order := range req.Order {
		groupReq.Order[i] = struct {
			Column int
			Dir    string
		}{
			Column: order.Column,
			Dir:    order.Dir,
		}
	}

	// Копируем Columns
	for i, col := range req.Columns {
		groupReq.Columns[i] = struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}{
			Data:       col.Data,
			Searchable: col.Searchable,
			Search: struct {
				Value string
			}{
				Value: col.Search.Value,
			},
		}
	}

	return groupReq
}

// ConvertToExchangeRepositoryRequest конвертирует DataTablesRequest в формат ExchangeRepository.
//
// Параметры:
//   - req: запрос из ParseDataTablesRequest
//
// Возвращает:
//   - *repositories.DataTablesRequest: запрос для ExchangeRepository (тот же тип, что и для групп)
func ConvertToExchangeRepositoryRequest(req *DataTablesRequest) *repositories.DataTablesRequest {
	exReq := &repositories.DataTablesRequest{
		Start:  req.Start,
		Length: req.Length,
		Search: req.Search,
		Order: make([]struct {
			Column int
			Dir    string
		}, len(req.Order)),
		Columns: make([]struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}, len(req.Columns)),
	}

	// Копируем Order
	for i, order := range req.Order {
		exReq.Order[i] = struct {
			Column int
			Dir    string
		}{
			Column: order.Column,
			Dir:    order.Dir,
		}
	}

	// Копируем Columns
	for i, col := range req.Columns {
		exReq.Columns[i] = struct {
			Data       string
			Searchable bool
			Search     struct {
				Value string
			}
		}{
			Data:       col.Data,
			Searchable: col.Searchable,
			Search: struct {
				Value string
			}{
				Value: col.Search.Value,
			},
		}
	}

	return exReq
}

// DataTablesResponse представляет стандартный ответ для DataTables.
//
// Формат ответа должен быть:
//
//	{
//	  "draw": 1,
//	  "recordsTotal": 100,
//	  "recordsFiltered": 50,
//	  "data": [...]
//	}
//
// Или в старом формате (для совместимости с PHP):
//
//	{
//	  "recordsTotal": 100,
//	  "recordsFiltered": 50,
//	  "aaData": [...]
//	}
type DataTablesResponse struct {
	Draw            int         `json:"draw,omitempty"`   // Номер запроса (для синхронизации)
	RecordsTotal    int         `json:"recordsTotal"`     // Общее количество записей
	RecordsFiltered int         `json:"recordsFiltered"`  // Количество записей после фильтрации
	Data            interface{} `json:"data,omitempty"`   // Данные (новый формат)
	AAData          interface{} `json:"aaData,omitempty"` // Данные (старый формат для совместимости с PHP)
}

// UserDataTablesRow представляет строку данных пользователя для DataTables.
//
// Формат соответствует PHP коду, который возвращает:
//   - chbx: пустая строка (для checkbox)
//   - DT_RowId: ID строки (например, "row_1")
//   - id, login, groups, active, name, email, create_date, modify_date, timestamp
type UserDataTablesRow struct {
	Chbx       string `json:"chbx"`        // Пустая строка для checkbox
	DTRowId    string `json:"DT_RowId"`    // ID строки (например, "row_1")
	ID         int    `json:"id"`          // ID пользователя
	Login      string `json:"login"`       // Логин
	Groups     string `json:"groups"`      // Список групп через запятую
	Active     string `json:"active"`      // "Active" или "Blocked"
	Name       string `json:"name"`        // Полное имя (LAST_NAME + NAME)
	Email      string `json:"email"`       // Email
	CreateDate string `json:"create_date"` // Форматированная дата создания
	ModifyDate string `json:"modify_date"` // Форматированная дата изменения
	Timestamp  string `json:"timestamp"`   // Форматированная дата последней активности
}

// GroupDataTablesRow представляет строку данных группы для DataTables.
//
// Формат соответствует PHP коду, который возвращает:
//   - chbx: пустая строка (для checkbox)
//   - DT_RowId: ID строки (например, "row_1")
//   - id, name, status, description
type GroupDataTablesRow struct {
	Chbx        string `json:"chbx"`        // Пустая строка для checkbox
	DTRowId     string `json:"DT_RowId"`    // ID строки (например, "row_1")
	ID          int    `json:"id"`          // ID группы
	Name        string `json:"name"`        // Название группы
	Status      string `json:"status"`      // "Active" или "Blocked"
	Description string `json:"description"` // Описание группы
}

// ConvertUserResponseToDataTablesFormat конвертирует ответ из UserRepository в формат DataTables.
//
// Параметры:
//   - draw: номер запроса (из DataTablesRequest.Draw)
//   - repoResponse: ответ из UserRepository.FindAllWithPagination
//
// Возвращает:
//   - *DataTablesResponse: ответ в формате DataTables
//
// Пример использования:
//
//	req := utils.ParseDataTablesRequest(c)
//	repoResponse, _ := userRepo.FindAllWithPagination(req)
//	response := utils.ConvertUserResponseToDataTablesFormat(req.Draw, repoResponse)
//	c.JSON(200, response)
func ConvertUserResponseToDataTablesFormat(draw int, repoResponse *repositories.UserDataTablesResponse) *DataTablesResponse {
	// Конвертируем UserDataTablesRow в формат DataTables
	aaData := make([]*UserDataTablesRow, len(repoResponse.Data))
	for i, row := range repoResponse.Data {
		aaData[i] = &UserDataTablesRow{
			Chbx:       "",                            // Пустая строка для checkbox
			DTRowId:    "row_" + strconv.Itoa(row.ID), // ID строки
			ID:         row.ID,
			Login:      row.Login,
			Groups:     row.Groups,
			Active:     row.Active,
			Name:       row.Name,
			Email:      row.Email,
			CreateDate: row.CreateDate,
			ModifyDate: row.ModifyDate,
			Timestamp:  row.Timestamp,
		}
	}

	return &DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    repoResponse.RecordsTotal,
		RecordsFiltered: repoResponse.RecordsFiltered,
		AAData:          aaData, // Используем старый формат для совместимости с PHP
	}
}

// ConvertGroupResponseToDataTablesFormat конвертирует ответ из GroupRepository в формат DataTables.
//
// Параметры:
//   - draw: номер запроса (из DataTablesRequest.Draw)
//   - repoResponse: ответ из GroupRepository.FindAllWithPagination
//
// Возвращает:
//   - *DataTablesResponse: ответ в формате DataTables
//
// Пример использования:
//
//	req := utils.ParseDataTablesRequest(c)
//	repoResponse, _ := groupRepo.FindAllWithPagination(req)
//	response := utils.ConvertGroupResponseToDataTablesFormat(req.Draw, repoResponse)
//	c.JSON(200, response)
func ConvertGroupResponseToDataTablesFormat(draw int, repoResponse *repositories.DataTablesResponse) *DataTablesResponse {
	// Конвертируем Group в формат DataTables
	aaData := make([]*GroupDataTablesRow, len(repoResponse.Data))
	for i, group := range repoResponse.Data {
		status := "Blocked"
		if group.Active {
			status = "Active"
		}

		aaData[i] = &GroupDataTablesRow{
			Chbx:        "",                              // Пустая строка для checkbox
			DTRowId:     "row_" + strconv.Itoa(group.ID), // ID строки
			ID:          group.ID,
			Name:        group.Name,
			Status:      status,
			Description: group.Description,
		}
	}

	return &DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    repoResponse.RecordsTotal,
		RecordsFiltered: repoResponse.RecordsFiltered,
		AAData:          aaData, // Используем старый формат для совместимости с PHP
	}
}

// ExchangeDataTablesRowDTO представляет строку биржи для DataTables (DTO слой).
type ExchangeDataTablesRowDTO struct {
	Chbx           string  `json:"chbx"`          // Пустая строка для checkbox
	DTRowId        string  `json:"DT_RowId"`      // ID строки (например, "row_1")
	ID             int     `json:"id"`            // ID биржи
	Name           string  `json:"name"`          // Название биржи
	Status         string  `json:"status"`        // "Active" или "Blocked"
	URL            string  `json:"url"`           // URL биржи
	BaseURL        string  `json:"base_url"`      // Base URL
	WebsocketURL   *string `json:"websocket_url"` // WebSocket URL
	ClassToFactory string  `json:"class"`         // Имя класса для фабрики
}

// ConvertExchangeResponseToDataTablesFormat конвертирует ответ из ExchangeRepository в формат DataTables.
//
// Параметры:
//   - draw: номер запроса (из DataTablesRequest.Draw)
//   - repoResponse: ответ из ExchangeRepository.FindAllWithPagination
//
// Возвращает:
//   - *DataTablesResponse: ответ в формате DataTables
func ConvertExchangeResponseToDataTablesFormat(draw int, repoResponse *repositories.ExchangeDataTablesResponse) *DataTablesResponse {
	aaData := make([]*ExchangeDataTablesRowDTO, len(repoResponse.Data))
	for i, ex := range repoResponse.Data {
		status := "Blocked"
		if strings.EqualFold(ex.Active, "Active") {
			status = "Active"
		}
		aaData[i] = &ExchangeDataTablesRowDTO{
			Chbx:           "",
			DTRowId:        "row_" + strconv.Itoa(ex.ID),
			ID:             ex.ID,
			Name:           ex.Name,
			Status:         status,
			URL:            ex.URL,
			BaseURL:        ex.BaseURL,
			WebsocketURL:   ex.WebsocketURL,
			ClassToFactory: ex.ClassToFactory,
		}
	}

	return &DataTablesResponse{
		Draw:            draw,
		RecordsTotal:    repoResponse.RecordsTotal,
		RecordsFiltered: repoResponse.RecordsFiltered,
		AAData:          aaData,
	}
}

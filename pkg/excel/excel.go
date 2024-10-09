package excel

import (
	"fmt"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xuri/excelize/v2"
	"os"
	"sync"
	"time"
)

const filename = "users.xlsx"

type Excel struct {
	log *logger.Logger
	mu  sync.Mutex
}

func NewExcel(log *logger.Logger) *Excel {
	return &Excel{log: log}
}

type Question struct {
	ID       int
	UserID   int
	Question string
}

func (e *Excel) GenerateUserResultsExcelFile(results []Question, username string) (string, error) {
	start := time.Now()

	f := excelize.NewFile()

	defer func() {
		if err := f.Close(); err != nil {
			e.log.Error("failed to close excel: %v", err)
		}
	}()

	sheetName := "Sheet1"
	f.NewSheet(sheetName)

	headers := map[string]string{
		"A1": "ID вопроса",
		"B1": "ID пользователя",
		"C1": "Вопрос",
	}

	for cell, value := range headers {
		f.SetCellValue(sheetName, cell, value)
	}

	for i, result := range results {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), result.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), result.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), result.Question)
	}

	filename := fmt.Sprintf("question_result.xlsx")
	err := f.SaveAs(filename)
	if err != nil {
		e.log.Error("failed to save file: %s", filename)
		return "", err
	}

	end := time.Since(start)
	e.log.Info("[%s] by [%s] Время генерации файла: %f", filename, username, end.Seconds())
	return filename, nil
}

func (e *Excel) GetExcelFile(fileName string) (*[]byte, error) {

	file, err := os.Open(fileName)
	if err != nil {
		e.log.Error("os.Open: failed to open file: %v", err)
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			e.log.Error("%v", err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		e.log.Error("file.Stat: failed to get file stat: %v", err)
		return nil, err
	}

	fileSize := fileInfo.Size()
	fileID := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: make([]byte, fileSize),
	}

	if _, err = file.Read(fileID.Bytes); err != nil {
		e.log.Error("file.Read: failed to get read file: %v", err)
		return nil, err
	}

	return &fileID.Bytes, nil
}

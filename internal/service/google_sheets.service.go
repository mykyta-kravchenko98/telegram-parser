package service

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mykyta-kravchenko98/telegram-parser/internal/config"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	service       *sheets.Service
	spreadsheetId string
	currentRow    int
	mutex         sync.Mutex
}

type GoogleCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

func NewGoogleSheetsService(cfg config.GoogleCredentials) (*GoogleSheetsService, error) {
	ctx := context.Background()

	// Загрузка учетных данных сервисного аккаунта из JSON
	credsJSON, err := json.Marshal(GoogleCredentials{
		Type:                    cfg.Type,
		ProjectID:               cfg.ProjectID,
		PrivateKeyID:            cfg.PrivateKeyID,
		PrivateKey:              cfg.PrivateKey,
		ClientEmail:             cfg.ClientEmail,
		ClientID:                cfg.ClientID,
		AuthURI:                 cfg.AuthURI,
		TokenURI:                cfg.TokenURI,
		AuthProviderX509CertURL: cfg.AuthProviderX509CertURL,
		ClientX509CertURL:       cfg.ClientX509CertURL,
	})

	// Получение конфигурации JWT
	config, err := google.JWTConfigFromJSON(credsJSON, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("JWTConfigFromJSON failed: %v", err)
	}

	// Создание клиента с доступом к Google Sheets
	client := config.Client(ctx)

	// Создание сервиса Google Sheets
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to create sheets service: %v", err)
	}

	return &GoogleSheetsService{
		service:       srv,
		spreadsheetId: cfg.SpreadsheetID,
		currentRow:    1,
	}, nil
}

func (s *GoogleSheetsService) CreateSheetIfNotExists(sheetName string) error {
	spreadsheet, err := s.service.Spreadsheets.Get(s.spreadsheetId).Do()
	if err != nil {
		return err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return nil
		}
	}

	addSheetRequest := sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title: sheetName,
		},
	}
	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AddSheet: &addSheetRequest},
		},
	}

	_, err = s.service.Spreadsheets.BatchUpdate(s.spreadsheetId, &batchUpdateRequest).Do()
	if err == nil {
		s.resetRowCounter()
	}
	return err
}

func (s *GoogleSheetsService) AppendRow(sheetName string, values []interface{}) error {
	rangeData := fmt.Sprintf("%s!A%d", sheetName, s.getCurrentRow())
	s.incrementRow()

	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         [][]interface{}{values},
	}

	_, err := s.service.Spreadsheets.Values.Append(s.spreadsheetId, rangeData, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	return nil
}

func (s *GoogleSheetsService) incrementRow() {
	s.mutex.Lock()
	s.currentRow++
	s.mutex.Unlock()
}

func (s *GoogleSheetsService) getCurrentRow() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.currentRow
}

func (s *GoogleSheetsService) resetRowCounter() {
	s.mutex.Lock()
	s.currentRow = 1
	s.mutex.Unlock()
}

func (s *GoogleSheetsService) GetSheetLink() string {
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit#gid=0", s.spreadsheetId)
}

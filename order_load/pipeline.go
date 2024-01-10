package order_load

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/ildomm/cc_sq_disbursement/database"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/fee_calculator"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultJobPause is the interval between each job run
	DefaultJobPause = time.Duration(60) * time.Second
	WaitingPath     = "../orders/waiting/"
	ImportedPath    = "../orders/imported/"
	FailedPath      = "../orders/failed/"
)

var csvHeaderInfo = []string{"id", "merchant_reference", "amount", "created_at"}

type pipeline struct {
	ctx      context.Context
	querier  database.Querier
	feeCalc  fee_calculator.FeeCalculator
	JobPause time.Duration // The amount of time to pause between each job run. This is exported so it can be overridden in tests
}

func NewPipeline(ctx context.Context, querier database.Querier) *pipeline {
	p := &pipeline{
		ctx:      ctx,
		querier:  querier,
		feeCalc:  *fee_calculator.NewFeeCalculator(ctx, querier),
		JobPause: DefaultJobPause,
	}

	return p
}

// Run starts the loading pipeline
func (pp *pipeline) Run() {
	log.Print("starting loading orders from CSV`s")

	for {
		select {
		case <-pp.ctx.Done():
			log.Print("stopping loading orders from CSV`s")
			return
		default:
			pp.importFiles()

			time.Sleep(pp.JobPause)
		}
	}
}

// importFiles imports all the files waiting to be loaded
// files will be sitting in the folder /orders/waiting
// the file will be moved to /orders/imported after being loaded
func (pp *pipeline) importFiles() {
	files, err := pp.detectFilesWaiting(WaitingPath)

	if err != nil {
		log.Print(err)
		return
	}

	log.Print("files to process: ", len(files))

	if len(files) > 0 {
		for _, file := range files {
			err := pp.importOrdersFromCSV(file)

			targetPath := ""
			if err != nil {
				log.Print("error importing orders from CSV:", err)

				// Error path
				targetPath = strings.Replace(file, WaitingPath, FailedPath, 1)
			} else {
				// Success path
				targetPath = strings.Replace(file, WaitingPath, ImportedPath, 1)
			}

			err = pp.moveFile(file, targetPath)
			if err != nil {
				log.Print("error moving file:", err)
			}
		}
	}
}

// importOrdersFromCSV imports orders from a CSV file
// the file will be ignored if it has an invalid format
// the file will be ignored if the merchant doesn't exist
// the file will be ignored if the order already exists
func (pp *pipeline) importOrdersFromCSV(filePath string) error {
	log.Printf("importing orders from CSV file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Variable to track if the current line is the first line (header)
	isFirstLine := true
	counter := 0

	for _, record := range records {
		// Skip the first line (header)
		if isFirstLine {
			isFirstLine = false
			continue
		}

		if len(record) != len(csvHeaderInfo) {
			return errors.New("invalid CSV format")
		}

		if counter%1000 == 0 {
			log.Printf("importing order %s. %d/%d", record[0], counter, len(records))
		}
		counter++

		order, err := pp.buildOrder(record)
		if err != nil {
			return err
		}

		existingOrder, err := pp.querier.SelectOrder(pp.ctx, order.ID)
		if err != nil {
			return fmt.Errorf("error checking if order exists: %v", err)
		}

		// Skip if the order already exists
		if existingOrder != nil {
			log.Printf("order already exists: %s", order.ID)
			continue
		}

		err = pp.querier.InsertOrder(pp.ctx, *order)
		if err != nil {
			return fmt.Errorf("error inserting order: %v", err)
		}
	}

	log.Printf("orders imported successfully from CSV file: %s", filePath)
	return nil
}

// buildOrder builds an order from a CSV record
// will return an error if the merchant doesn't exist
// will return an error if the amount is not a valid float
// will return an error if the created_at is not a valid date
func (pp *pipeline) buildOrder(record []string) (*entities.Order, error) {

	orderID := record[0]
	merchantReference := record[1]
	amount := record[2]
	createdAtStr := record[3]

	merchant, err := pp.querier.SelectMerchantByReference(pp.ctx, merchantReference)
	if err != nil {
		return nil, fmt.Errorf("error checking if merchant exists: %v", err)
	}

	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting amount to float: %v", err)
	}

	createdAt, err := time.Parse(time.DateOnly, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %v", err)
	}

	order := entities.Order{
		ID:         orderID,
		MerchantID: merchant.ID,
		Amount:     amountFloat,
		CreatedAt:  createdAt,
	}

	// calculates the fee amount based on the order amount and fee percentages.
	order.FeeAmount = pp.feeCalc.CalculateFeeAmount(order.Amount)

	return &order, nil
}

// moveFile moves a file from one path to another
func (pp *pipeline) moveFile(filePath, targetPath string) error {
	err := os.Rename(filePath, targetPath)
	if err != nil {
		return err
	}
	return nil
}

// detectFilesWaiting Detects if there are files waiting to be processed
// files will be sitting in the folder /orders/waiting
func (pp *pipeline) detectFilesWaiting(folderPath string) ([]string, error) {
	log.Print("detecting files...")

	var files []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file (not a directory)
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

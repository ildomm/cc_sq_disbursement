package order_process

import (
	"context"
	"github.com/ildomm/cc_sq_disbursement/database"
	"github.com/ildomm/cc_sq_disbursement/entities"
	"github.com/ildomm/cc_sq_disbursement/fee_calculator"
	"github.com/ildomm/cc_sq_disbursement/system"
	"log"
	"time"
)

type pipeline struct {
	ctx     context.Context
	querier database.Querier
	feeCalc fee_calculator.FeeCalculator
}

func NewPipeline(ctx context.Context, querier database.Querier) *pipeline {
	p := &pipeline{
		ctx:     ctx,
		querier: querier,
		feeCalc: *fee_calculator.NewFeeCalculator(ctx, querier),
	}

	return p
}

// Run starts the processing pipeline
func (pp *pipeline) Run(day time.Time) {
	log.Printf("start processing orders from day %s", day)
	pp.process(day)
	log.Printf("finish processing orders from day %s", day)
}

func (pp *pipeline) process(day time.Time) {

	// Create daily disbursements
	err := pp.dailyDisbursements(day)
	if err != nil {
		log.Printf("error creating daily disbursements: %s", err)
		return
	}

	// Create weekly disbursements
	err = pp.weeklyDisbursements(day)
	if err != nil {
		log.Printf("error creating weekly disbursements: %s", err)
		return
	}

	// Create monthly disbursements
	err = pp.monthlyDisbursements(day)
	if err != nil {
		log.Printf("error creating monthly disbursements: %s", err)
		return
	}
}

// dailyDisbursements creates the daily disbursements for the day
func (pp *pipeline) dailyDisbursements(day time.Time) error {
	// TODO: Implement in a transaction
	// Not doing right now because we are running out of time
	// TODO: Start transaction

	disbursements, err := pp.querier.SelectSumOrders(pp.ctx, day)
	if err != nil {
		return err
	}
	for _, disbursement := range disbursements {

		// Check if it is necessary to correct the fee amount
		feeAmountCorrection, err := pp.feeCalc.CalculateFeeAmountCorrection(day, disbursement)
		if err != nil {
			return err
		}
		disbursement.FeeAmountCorrection = feeAmountCorrection

		err = pp.querier.InsertDisbursement(pp.ctx, disbursement)
		if err != nil {
			return err
		}
	}

	if len(disbursements) > 0 {
		// Mark orders as disbursed
		err = pp.querier.MarkOrdersAsDisbursed(pp.ctx, day)
		if err != nil {
			return err
		}
	}

	// TODO: END transaction

	return nil
}

// weeklyDisbursements creates the weekly disbursements for the week
// It is created on Mondays
// It is calculated by summing the orders for the last week
func (pp *pipeline) weeklyDisbursements(day time.Time) error {
	if day.Weekday() != time.Monday {
		return nil
	}

	// Last week, with Monday as the first day of the week
	firstDayLastWeek := day.AddDate(0, 0, -7)
	lastDayLastWeek := day.AddDate(0, 0, -1)

	// Get the sum of orders for the last week
	disbursements, err := pp.querier.SelectSumDisbursements(pp.ctx,
		firstDayLastWeek,
		lastDayLastWeek,
		entities.WeeklyDisbursementFrequency)
	if err != nil {
		return err
	}

	for _, disbursement := range disbursements {
		err = pp.querier.InsertDisbursement(pp.ctx, disbursement)
		if err != nil {
			return err
		}
	}

	return nil
}

// monthlyDisbursements creates the monthly disbursements for the month
// It is created on the first day of the month
// It is calculated by summing the orders for the last month
func (pp *pipeline) monthlyDisbursements(day time.Time) error {
	if day.Day() != 1 {
		return nil
	}

	// Get the sum of orders for the last week
	disbursements, err := pp.querier.SelectSumDisbursements(pp.ctx,
		system.FirstDayOfLastMonth(day),
		system.LastDayOfLastMonth(day),
		entities.MonthlyDisbursementFrequency)
	if err != nil {
		return err
	}

	for _, disbursement := range disbursements {
		err = pp.querier.InsertDisbursement(pp.ctx, disbursement)
		if err != nil {
			return err
		}
	}

	return nil
}

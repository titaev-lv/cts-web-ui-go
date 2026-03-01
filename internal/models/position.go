package models

import "time"

type PositionSummary struct {
	PositionID       int
	ContractName     string
	ExchangeName     string
	MarketType       string
	Status           string
	Created          *time.Time
	Closed           *time.Time
	FinalPosition    *float64
	FinalAvgPrice    *float64
	FeeBaseTotal     *float64
	FeeTotal         *float64
	FundingTotal     *float64
	TotalRealizedPnL *float64
}

type PositionDetail struct {
	PositionID       int
	ContractName     string
	ExchangeName     string
	MarketType       string
	Status           string
	Created          *time.Time
	Closed           *time.Time
	FinalPosition    *float64
	FinalAvgPrice    *float64
	FeeBaseTotal     *float64
	FeeTotal         *float64
	FundingTotal     *float64
	TotalRealizedPnL *float64
	TransCount       int
}

type PositionTransaction struct {
	ID        int
	Type      string
	Price     float64
	Volume    float64
	FeeBase   float64
	Fee       float64
	Funding   float64
	TransDate *time.Time
}

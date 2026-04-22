package backtest

import (
	"context"
	"fmt"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// WalkForwardFold represents a single fold in walk-forward analysis.
type WalkForwardFold struct {
	FoldIndex   int     `json:"foldIndex"`
	TrainStart  int     `json:"trainStart"`
	TrainEnd    int     `json:"trainEnd"`
	TestStart   int     `json:"testStart"`
	TestEnd     int     `json:"testEnd"`
	TrainReturn float64 `json:"trainReturn"`
	TestReturn  float64 `json:"testReturn"`
	Robustness  float64 `json:"robustness"`
}

// WalkForwardResult holds the complete walk-forward analysis.
type WalkForwardResult struct {
	Strategy      string            `json:"strategy"`
	Symbol        string            `json:"symbol"`
	Period        string            `json:"period"`
	NSplits       int               `json:"nSplits"`
	TrainRatio    float64           `json:"trainRatio"`
	AverageTrain  float64           `json:"averageTrain"`
	AverageTest   float64           `json:"averageTest"`
	AvgRobustness float64           `json:"avgRobustness"`
	Verdict       string            `json:"verdict"`
	Folds         []WalkForwardFold `json:"folds"`
}

// RunWalkForwardBacktest performs walk-forward analysis on a strategy.
// Splits data into nSplits folds, trains on trainRatio, tests on remainder.
// Computes robustness = testReturn / trainReturn.
// Verdict: >=0.8 "ROBUST", >=0.5 "MODERATE", >=0.2 "WEAK", <0.2 "OVERFITTED"
func RunWalkForwardBacktest(cfg *config.Config, symbol, strategy, period, interval string, initialCapital, commissionPct, slippagePct float64, nSplits int, trainRatio float64) error {
	// Validate inputs
	if err := utils.ValidateStrategy(strategy); err != nil {
		return err
	}
	if err := utils.ValidatePeriod(period); err != nil {
		return err
	}
	if err := utils.ValidateInterval(interval); err != nil {
		return err
	}

	// Validate nSplits and trainRatio
	if nSplits < 2 {
		return fmt.Errorf("nSplits must be >= 2")
	}
	if trainRatio <= 0 || trainRatio >= 1 {
		return fmt.Errorf("trainRatio must be between 0 and 1")
	}

	// Create HTTP and Yahoo clients
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	// Fetch all candles for the period
	ctx := context.Background()
	candles, err := yahooClient.GetChart(ctx, symbol, interval, period)
	if err != nil {
		return fmt.Errorf("fetch chart: %w", err)
	}

	if len(candles) == 0 {
		return fmt.Errorf("no candles retrieved for %s", symbol)
	}

	// Get strategy
	strat := GetStrategy(strategy)
	if strat == nil {
		return fmt.Errorf("unknown strategy: %s", strategy)
	}

	// Calculate fold size
	foldSize := len(candles) / nSplits
	if foldSize == 0 {
		return fmt.Errorf("not enough candles for %d splits", nSplits)
	}

	result := WalkForwardResult{
		Strategy:   strategy,
		Symbol:     symbol,
		Period:     period,
		NSplits:    nSplits,
		TrainRatio: trainRatio,
		Folds:      make([]WalkForwardFold, 0),
	}

	var trainReturns, testReturns, robustnesses []float64

	// Execute walk-forward analysis
	for fold := 0; fold < nSplits; fold++ {
		foldStart := fold * foldSize
		var foldEnd int
		if fold == nSplits-1 {
			// Last fold gets all remaining data
			foldEnd = len(candles)
		} else {
			foldEnd = foldStart + foldSize
		}

		trainEnd := foldStart + int(float64(foldEnd-foldStart)*trainRatio)

		// Extract train and test data
		trainCandles := candles[foldStart:trainEnd]
		testCandles := candles[trainEnd:foldEnd]

		// Skip if either set is empty
		if len(trainCandles) == 0 || len(testCandles) == 0 {
			continue
		}

		// Run backtest on train data
		trainResult := RunBacktest(symbol, strategy, trainCandles, strat, initialCapital, commissionPct, slippagePct, false, false)
		trainReturn := trainResult.TotalReturn

		// Run backtest on test data
		testResult := RunBacktest(symbol, strategy, testCandles, strat, initialCapital, commissionPct, slippagePct, false, false)
		testReturn := testResult.TotalReturn

		// Calculate robustness
		robustness := 0.0
		if trainReturn > 0 {
			robustness = testReturn / trainReturn
		} else if trainReturn < 0 {
			// If training was negative, robustness = testReturn / trainReturn (both negative)
			robustness = testReturn / trainReturn
		}
		// If trainReturn == 0, robustness stays 0

		trainReturns = append(trainReturns, trainReturn)
		testReturns = append(testReturns, testReturn)
		robustnesses = append(robustnesses, robustness)

		// Record fold
		foldRecord := WalkForwardFold{
			FoldIndex:   fold,
			TrainStart:  foldStart,
			TrainEnd:    trainEnd,
			TestStart:   trainEnd,
			TestEnd:     foldEnd,
			TrainReturn: trainReturn,
			TestReturn:  testReturn,
			Robustness:  robustness,
		}
		result.Folds = append(result.Folds, foldRecord)
	}

	// Calculate averages
	if len(trainReturns) > 0 {
		trainSum := 0.0
		testSum := 0.0
		robustSum := 0.0

		for i := range trainReturns {
			trainSum += trainReturns[i]
			testSum += testReturns[i]
			robustSum += robustnesses[i]
		}

		result.AverageTrain = trainSum / float64(len(trainReturns))
		result.AverageTest = testSum / float64(len(testReturns))
		result.AvgRobustness = robustSum / float64(len(robustnesses))
	}

	// Determine verdict
	avgRobustness := result.AvgRobustness
	if avgRobustness >= 0.8 {
		result.Verdict = "ROBUST"
	} else if avgRobustness >= 0.5 {
		result.Verdict = "MODERATE"
	} else if avgRobustness >= 0.2 {
		result.Verdict = "WEAK"
	} else {
		result.Verdict = "OVERFITTED"
	}

	// Print result as JSON
	if err := utils.PrintJSON(result); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}

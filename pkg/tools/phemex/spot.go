package phemex

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// PHEMEX spot endpoints use 's' prefix on symbols (sBTCUSDT) and Ev/Ep-scaled
// numeric fields (multiply real value by 1e8 to produce the integer wire value).
// Reference: https://github.com/phemex/phemex-api-docs/blob/master/Public-Spot-API-en.md

// toSpotSymbol ensures the 's' prefix on a spot symbol.
// "BTCUSDT" -> "sBTCUSDT". "sBTCUSDT" -> "sBTCUSDT" (idempotent).
func toSpotSymbol(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "s") || strings.HasPrefix(s, "S") {
		return "s" + strings.TrimPrefix(strings.TrimPrefix(s, "s"), "S")
	}
	return "s" + s
}

// scaleEv converts a real decimal value to a PHEMEX spot Ev/Ep wire string
// by multiplying by 1e8 and formatting as an integer. Negative values are
// clamped to 0 (PHEMEX expects non-negative quantities/prices).
func scaleEv(v float64) string {
	if v <= 0 {
		return "0"
	}
	scaled := int64(v*1e8 + 0.5) // round to nearest
	return strconv.FormatInt(scaled, 10)
}

// RunSpotBalance: GET /spot/wallets?currency=USDT
func RunSpotBalance(cfg *config.Config, currency string) error {
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	if currency == "" {
		currency = "USDT"
	}
	body, err := pc.Do(c, "GET", "/spot/wallets",
		map[string]string{"currency": strings.ToUpper(currency)}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s spot\n", pc.BaseURL())
	return printRaw(body)
}

// RunSpotOrders: GET /spot/orders?symbol=sBTCUSDT
// Returns active spot orders for the symbol. The caller passes the symbol
// without the 's' prefix; we add it automatically.
func RunSpotOrders(cfg *config.Config, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("--symbol is required (e.g. BTCUSDT)")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	body, err := pc.Do(c, "GET", "/spot/orders",
		map[string]string{"symbol": toSpotSymbol(strings.ToUpper(symbol))}, nil)
	if err != nil {
		// PHEMEX returns code 10002 OM_ORDER_NOT_FOUND when zero active orders.
		// Translate to empty list for ergonomic scripting.
		if strings.Contains(err.Error(), "10002") || strings.Contains(string(body), "OM_ORDER_NOT_FOUND") {
			fmt.Fprintf(printToStderr(), "# base=%s spot (no active orders)\n", pc.BaseURL())
			return utils.PrintJSON(map[string]interface{}{"code": 0, "data": map[string]interface{}{"rows": []interface{}{}}})
		}
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s spot\n", pc.BaseURL())
	return printRaw(body)
}

// SpotPlaceOpts describes a spot order (entry + optional auto SL/TP follow-ups).
type SpotPlaceOpts struct {
	Symbol      string  // e.g. "BTCUSDT" — no 's' prefix needed, we add it
	Side        string  // "Buy" or "Sell"
	OrdType     string  // "Market" or "Limit"
	QuoteQty    float64 // USDT amount (preferred for Buy); 0 if unused
	BaseQty     float64 // base asset units (preferred for Sell); 0 if unused
	Price       float64 // for Limit orders
	SL          float64 // optional stop-loss trigger price (auto-place Stop Sell)
	TP          float64 // optional take-profit price (auto-place Limit Sell)
	TimeInForce string  // optional, default GoodTillCancel
	ClOrdID     string  // optional client order ID
}

// RunSpotPlace places the main spot order. If SL or TP are provided AND the
// main order side is Buy, it immediately follows up with the conditional
// orders (Stop Sell for SL, Limit Sell for TP) using the same base qty.
// Emits a combined JSON {main, stopLoss, takeProfit} on stdout.
func RunSpotPlace(cfg *config.Config, o SpotPlaceOpts) error {
	if o.Symbol == "" || o.Side == "" || o.OrdType == "" {
		return fmt.Errorf("--symbol --side --ord-type are required")
	}
	if o.QuoteQty == 0 && o.BaseQty == 0 {
		return fmt.Errorf("either --quote-qty or --base-qty must be provided")
	}
	if o.OrdType == "Limit" && o.Price == 0 {
		return fmt.Errorf("--price is required for Limit orders")
	}
	if o.Side != "Buy" && o.Side != "Sell" {
		return fmt.Errorf("--side must be Buy or Sell")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}

	apiSymbol := toSpotSymbol(strings.ToUpper(o.Symbol))

	// Build main order query
	q := map[string]string{
		"symbol":  apiSymbol,
		"side":    o.Side,
		"ordType": o.OrdType,
	}
	if o.QuoteQty > 0 {
		q["qtyType"] = "ByQuote"
		q["quoteQtyEv"] = scaleEv(o.QuoteQty)
	} else {
		q["qtyType"] = "ByBase"
		q["baseQtyEv"] = scaleEv(o.BaseQty)
	}
	if o.Price > 0 {
		q["priceEp"] = scaleEv(o.Price)
	}
	if o.TimeInForce != "" {
		q["timeInForce"] = o.TimeInForce
	}
	if o.ClOrdID != "" {
		q["clOrdID"] = o.ClOrdID
	}

	c, cancel := ctx()
	defer cancel()

	// For Buy with auto SL/TP, snapshot base asset balance BEFORE so we can
	// compute the actual fill qty via wallet diff (more reliable than parsing
	// the async Market order response).
	baseCurrency := baseCurrencyFromSymbol(apiSymbol)
	var balanceBefore float64
	if o.Side == "Buy" && (o.SL > 0 || o.TP > 0) {
		balanceBefore = fetchSpotBalance(pc, baseCurrency)
	}

	mainBody, mainErr := pc.Do(c, "PUT", "/spot/orders/create", q, nil)

	result := map[string]interface{}{
		"main":       jsonOrRaw(mainBody),
		"stopLoss":   nil,
		"takeProfit": nil,
	}
	if mainErr != nil {
		fmt.Fprintf(printToStderr(), "# base=%s spot (main order failed)\n", pc.BaseURL())
		_ = utils.PrintJSON(result)
		return mainErr
	}

	// Only auto-place SL/TP for Buy orders (we just opened a position).
	if o.Side == "Buy" && (o.SL > 0 || o.TP > 0) {
		// Best signal of how much base asset was bought is the wallet delta.
		// Market orders fill async on PHEMEX testnet within ~1s; poll the wallet.
		filledQty := extractFilledBaseQty(mainBody)
		if filledQty == 0 {
			for i := 0; i < 4; i++ {
				time.Sleep(500 * time.Millisecond)
				after := fetchSpotBalance(pc, baseCurrency)
				delta := after - balanceBefore
				if delta > 0 {
					filledQty = delta
					break
				}
			}
		}
		if filledQty == 0 && o.BaseQty > 0 {
			filledQty = o.BaseQty
		}
		if filledQty == 0 && o.QuoteQty > 0 && o.Price > 0 {
			filledQty = o.QuoteQty / o.Price
		}
		// Note: if still unknown, SL/TP follow-ups skip with note in JSON.

		if filledQty > 0 {
			if o.SL > 0 {
				slQuery := map[string]string{
					"symbol":        apiSymbol,
					"side":          "Sell",
					"ordType":       "Stop",
					"qtyType":       "ByBase",
					"baseQtyEv":     scaleEv(filledQty),
					"stopPxEp":      scaleEv(o.SL),
					"trigger":       "ByLastPrice",
					"stopDirection": "Falling", // fire when last price falls to/through stopPxEp
				}
				slBody, slErr := pc.Do(c, "PUT", "/spot/orders/create", slQuery, nil)
				if slErr != nil {
					result["stopLoss"] = map[string]interface{}{"error": slErr.Error(), "body": jsonOrRaw(slBody)}
				} else {
					result["stopLoss"] = jsonOrRaw(slBody)
				}
			}
			if o.TP > 0 {
				tpQuery := map[string]string{
					"symbol":    apiSymbol,
					"side":      "Sell",
					"ordType":   "Limit",
					"qtyType":   "ByBase",
					"baseQtyEv": scaleEv(filledQty),
					"priceEp":   scaleEv(o.TP),
					"timeInForce": "GoodTillCancel",
				}
				tpBody, tpErr := pc.Do(c, "PUT", "/spot/orders/create", tpQuery, nil)
				if tpErr != nil {
					result["takeProfit"] = map[string]interface{}{"error": tpErr.Error(), "body": jsonOrRaw(tpBody)}
				} else {
					result["takeProfit"] = jsonOrRaw(tpBody)
				}
			}
		} else {
			result["note"] = "SL/TP skipped: could not determine filled base qty from main order response"
		}
	}

	fmt.Fprintf(printToStderr(), "# base=%s spot\n", pc.BaseURL())
	return utils.PrintJSON(result)
}

// RunSpotCancel: DELETE /spot/orders?symbol=sBTCUSDT&orderID=...
func RunSpotCancel(cfg *config.Config, symbol, orderID string) error {
	if symbol == "" || orderID == "" {
		return fmt.Errorf("--symbol and --order-id are required")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	body, err := pc.Do(c, "DELETE", "/spot/orders", map[string]string{
		"symbol":  toSpotSymbol(strings.ToUpper(symbol)),
		"orderID": orderID,
	}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s spot\n", pc.BaseURL())
	return printRaw(body)
}

// jsonOrRaw decodes raw bytes to a JSON-friendly value; falls back to string.
func jsonOrRaw(b []byte) interface{} {
	var v interface{}
	if err := json.Unmarshal(b, &v); err == nil {
		return v
	}
	return string(b)
}

// fetchSpotBalance reads the base asset (e.g. BTC) balance from /spot/wallets.
// Returns the real balance (already divided by 1e8). 0 if not found.
func fetchSpotBalance(pc *client.PhemexPrivateClient, currency string) float64 {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	body, err := pc.Do(c, "GET", "/spot/wallets",
		map[string]string{"currency": strings.ToUpper(currency)}, nil)
	if err != nil {
		return 0
	}
	var env struct {
		Data []struct {
			BalanceEv json.Number `json:"balanceEv"`
			Currency  string      `json:"currency"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &env) != nil {
		return 0
	}
	for _, w := range env.Data {
		if strings.EqualFold(w.Currency, currency) {
			if f, err := w.BalanceEv.Float64(); err == nil {
				return f / 1e8
			}
		}
	}
	return 0
}

// baseCurrencyFromSymbol extracts the base asset from a PHEMEX spot symbol.
// "sBTCUSDT" -> "BTC". Falls back by trimming "USDT" suffix.
func baseCurrencyFromSymbol(apiSymbol string) string {
	s := strings.TrimPrefix(apiSymbol, "s")
	s = strings.TrimPrefix(s, "S")
	if strings.HasSuffix(s, "USDT") {
		return strings.TrimSuffix(s, "USDT")
	}
	if strings.HasSuffix(s, "USDC") {
		return strings.TrimSuffix(s, "USDC")
	}
	return s
}

// extractFilledBaseQty looks for cumBaseQtyEv in the main order response and
// divides by 1e8 to recover the real qty. Returns 0 if not present.
func extractFilledBaseQty(body []byte) float64 {
	var env struct {
		Data struct {
			CumBaseQtyEv json.Number `json:"cumBaseQtyEv"`
			BaseQtyEv    json.Number `json:"baseQtyEv"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return 0
	}
	if env.Data.CumBaseQtyEv != "" {
		if f, err := env.Data.CumBaseQtyEv.Float64(); err == nil && f > 0 {
			return f / 1e8
		}
	}
	if env.Data.BaseQtyEv != "" {
		if f, err := env.Data.BaseQtyEv.Float64(); err == nil && f > 0 {
			return f / 1e8
		}
	}
	return 0
}

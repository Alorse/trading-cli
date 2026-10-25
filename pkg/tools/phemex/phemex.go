// Package phemex implements authenticated PHEMEX subcommands:
// account, positions, orders, place-order, cancel.
// All operations default to testnet via PHEMEX_API_URL env var.
package phemex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// newClient builds a private PHEMEX client from config.
func newClient(cfg *config.Config) (*client.PhemexPrivateClient, error) {
	if cfg.PhemexAPIID == "" {
		return nil, fmt.Errorf("PHEMEX_API_ID is not set (load via .env or env var)")
	}
	if cfg.PhemexAPISecret == "" {
		return nil, fmt.Errorf("PHEMEX_API_SECRET is not set")
	}
	return client.NewPhemexPrivateClient(cfg.PhemexAPIURL, cfg.PhemexAPIID, cfg.PhemexAPISecret), nil
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// printRaw prints a pretty-printed JSON envelope to stdout.
func printRaw(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		fmt.Println(string(b))
		return nil
	}
	return utils.PrintJSON(v)
}

// RunAccount: GET /g-accounts/accountPositions?currency=USDT
// Returns the full account + positions snapshot.
func RunAccount(cfg *config.Config, currency string) error {
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	if currency == "" {
		currency = "USDT"
	}
	body, err := pc.Do(c, "GET", "/g-accounts/accountPositions",
		map[string]string{"currency": strings.ToUpper(currency)}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s\n", pc.BaseURL())
	return printRaw(body)
}

// RunPositions: GET /g-accounts/positions?currency=USDT
// Returns positions only (lighter than account).
func RunPositions(cfg *config.Config, currency string) error {
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	if currency == "" {
		currency = "USDT"
	}
	body, err := pc.Do(c, "GET", "/g-accounts/positions",
		map[string]string{"currency": strings.ToUpper(currency)}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s\n", pc.BaseURL())
	return printRaw(body)
}

// RunActiveOrders: GET /g-orders/activeList?symbol=BTCUSDT
func RunActiveOrders(cfg *config.Config, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("--symbol is required (e.g. BTCUSDT)")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	body, err := pc.Do(c, "GET", "/g-orders/activeList",
		map[string]string{"symbol": strings.ToUpper(symbol)}, nil)
	if err != nil {
		// PHEMEX returns code 10002 "OM_ORDER_NOT_FOUND" when there are zero active orders.
		// Translate that to an empty-list success response for ergonomic scripting.
		if strings.Contains(err.Error(), "10002") || strings.Contains(string(body), "OM_ORDER_NOT_FOUND") {
			fmt.Fprintf(printToStderr(), "# base=%s (no active orders)\n", pc.BaseURL())
			return utils.PrintJSON(map[string]interface{}{"code": 0, "data": map[string]interface{}{"rows": []interface{}{}}})
		}
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s\n", pc.BaseURL())
	return printRaw(body)
}

// PlaceOrderOpts describes a USDT-perpetual order.
type PlaceOrderOpts struct {
	Symbol      string // e.g. "BTCUSDT"
	Side        string // "Buy" | "Sell"
	PosSide     string // "Long" | "Short" | "Merged"
	OrdType     string // "Market" | "Limit"
	OrderQtyRq  string // contract quantity (base asset units, string per PHEMEX)
	PriceRp     string // price (limit orders) — empty for market
	StopLossRp  string // optional
	TPRp        string // optional take profit
	TimeInForce string // optional, default GoodTillCancel
	ReduceOnly  bool
	ClOrdID     string // optional client order id
}

// RunPlaceOrder: PUT /g-orders/create with parameters in the QUERY string.
func RunPlaceOrder(cfg *config.Config, o PlaceOrderOpts) error {
	if o.Symbol == "" || o.Side == "" || o.PosSide == "" || o.OrdType == "" || o.OrderQtyRq == "" {
		return fmt.Errorf("--symbol --side --pos-side --ord-type --qty are all required")
	}
	o.Symbol = strings.ToUpper(o.Symbol)
	if o.OrdType == "Limit" && o.PriceRp == "" {
		return fmt.Errorf("--price is required for Limit orders")
	}

	pc, err := newClient(cfg)
	if err != nil {
		return err
	}

	q := map[string]string{
		"symbol":     o.Symbol,
		"side":       o.Side,
		"posSide":    o.PosSide,
		"ordType":    o.OrdType,
		"orderQtyRq": o.OrderQtyRq,
	}
	if o.PriceRp != "" {
		q["priceRp"] = o.PriceRp
	}
	if o.StopLossRp != "" {
		q["stopLossRp"] = o.StopLossRp
	}
	if o.TPRp != "" {
		q["takeProfitRp"] = o.TPRp
	}
	if o.TimeInForce != "" {
		q["timeInForce"] = o.TimeInForce
	}
	if o.ReduceOnly {
		q["reduceOnly"] = "true"
	}
	if o.ClOrdID != "" {
		q["clOrdID"] = o.ClOrdID
	}

	c, cancel := ctx()
	defer cancel()
	body, err := pc.Do(c, "PUT", "/g-orders/create", q, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s\n", pc.BaseURL())
	return printRaw(body)
}

// RunSetLeverage: PUT /g-positions/leverage
// Sets long+short leverage caps on a symbol in hedged mode.
// marginType: "cross" (default) or "isolated". leverage is positive integer 1-100.
// PHEMEX convention: cross = negative leverageRr, isolated = positive.
func RunSetLeverage(cfg *config.Config, symbol string, leverage int, marginType string) error {
	if symbol == "" {
		return fmt.Errorf("--symbol is required")
	}
	if leverage < 1 || leverage > 100 {
		return fmt.Errorf("--leverage must be between 1 and 100")
	}
	mt := strings.ToLower(marginType)
	if mt == "" {
		mt = "cross"
	}
	signed := leverage
	if mt == "cross" {
		signed = -leverage
	} else if mt != "isolated" {
		return fmt.Errorf("--margin must be 'cross' or 'isolated'")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	lev := fmt.Sprintf("%d", signed)
	body, err := pc.Do(c, "PUT", "/g-positions/leverage", map[string]string{
		"symbol":          strings.ToUpper(symbol),
		"longLeverageRr":  lev,
		"shortLeverageRr": lev,
	}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s leverage=%s (negative=cross)\n", pc.BaseURL(), lev)
	return printRaw(body)
}

// RunCancelOrder: DELETE /g-orders/cancel?orderID=...&symbol=...&posSide=...
func RunCancelOrder(cfg *config.Config, symbol, orderID, posSide string) error {
	if symbol == "" || orderID == "" || posSide == "" {
		return fmt.Errorf("--symbol --order-id --pos-side are all required")
	}
	pc, err := newClient(cfg)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	body, err := pc.Do(c, "DELETE", "/g-orders/cancel", map[string]string{
		"orderID": orderID,
		"symbol":  strings.ToUpper(symbol),
		"posSide": posSide,
	}, nil)
	if err != nil {
		fmt.Fprintf(printToStderr(), "%s\n", string(body))
		return err
	}
	fmt.Fprintf(printToStderr(), "# base=%s\n", pc.BaseURL())
	return printRaw(body)
}

# trading-cli — agent memory

Writable scratchpad for your own use across hourly runs. Edit freely.
Default sections below are suggestions — adapt, rename, or drop sections as needed. Just don't lose the file.

## Hourly run log

Append one entry per run. Newest at the top. Trim entries older than 30 days when the file grows past ~200 lines.

Format suggestion:
```
### YYYY-MM-DD HH:MM UTC
- Balance: <N> USDT | Open: <N>/5 | Used margin: <N>
- Action: <opened X | closed Y | no action because Z>
- Reasoning: <one or two sentences>
```

(no runs yet)

## Active trades

Track positions you opened, with the rationale and current status. Move to "Closed trades" on exit.

Format suggestion:
```
- BTCUSDT Long | orderID <...> | clOrdID <...> | opened YYYY-MM-DD HH:MM
  entry: 77556.20 | SL: 76000 | TP: 79000
  reason: scalp-scan score 78 + funding -0.05% + MTF MOSTLY BULLISH
```

- BTCUSDT Long | orderID 9a6ebc87-70d4-4e31-bb9d-2a5e586fcab8 | clOrdID picoclaw-test-001
  entry: 77556.20 | SL: 76000 | TP: 79000
  reason: TEST trade placed manually by Claude during setup. Leave alone unless user says otherwise.

## Closed trades

Record outcomes for pattern analysis. Note what worked vs what didn't.

```
- BTCUSDT Long | opened YYYY-MM-DD HH:MM | closed YYYY-MM-DD HH:MM
  result: +1.8% | PnL: +9 USDT | exit: TP1
  lesson: <what to repeat / avoid>
```

(none yet)

## Market regime notes

Higher-level context that affects all decisions this week.

Examples:
- "BTC ranging $77k-$80k since Monday, no clear breakout yet"
- "Alts pumping on Asia hours (00:00-04:00 UTC), quiet during US"
- "Funding broadly negative across majors → bullish bias for longs"

(empty)

## Rejected setups

Symbols you considered but didn't take. Patterns emerge over time.

```
- YYYY-MM-DD HH:MM SYMBOL — score X, rejected because <reason>
```

(empty)

## User preferences / overrides

Things the user has explicitly asked once that you should remember.

Examples:
- "User dislikes meme tokens (DOGE, PEPE, SHIB, etc.) — don't suggest them"
- "User wants to be notified before any close > $20 PnL"

(empty)

## Lessons learned

Distill what works and what doesn't, in the user's actual account, over time.

(empty)

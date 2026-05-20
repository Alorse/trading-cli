# trading-cli — agent memory

Writable scratchpad for your own use across hourly runs. Edit freely.
Default sections below are suggestions — adapt as needed.

## State

Read this section at the start of every hourly run. Update at the end.

```
consecutiveEmptyHours: 0
lastAction: 2026-05-20 19:35 UTC — opened BTCUSDT Long (test trade, leave open)
regimeFlag: none
regimeFlagSince: -
regimeFlagReason: -
```

Apply scan parameters per `consecutiveEmptyHours` table in SKILL.md "Stale-market protocol".
Reset `consecutiveEmptyHours` to 0 on any trade opened. Increment by 1 on a run that opens nothing.

## Recent rejections (last 3 hours)

Append when you reject a candidate in step 8 of the decision tree. Prune entries older than 3 hours each run.

Format:
```
- SYMBOL @ YYYY-MM-DD HH:MM UTC — score N — reason: <short text>
```

Next hour, skip deep MTF analysis for a symbol in this list with same reason UNLESS its current score improved by ≥5 points.

(empty)

## Hourly run log

One entry per run. Newest at the top. Trim entries older than 30 days when the file grows past ~200 lines.

Format:
```
### YYYY-MM-DD HH:MM UTC
- Balance: <N> USDT | Open: <N>/5 | Used margin: <N>
- Scan params: min-score <S> min-turnover <T> [extra lens: <X>]
- Action: <opened X | closed Y | no action because Z>
- Reasoning: <one or two sentences>
```

(no runs yet)

## Active trades

Track positions you opened, with rationale and current status. Move to "Closed trades" on exit.

Format:
```
- BTCUSDT Long | orderID <...> | clOrdID <...> | opened YYYY-MM-DD HH:MM
  entry: <price> | SL: <price> | TP: <price>
  reason: <why you opened this>
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

Higher-level context that affects all decisions this week. Update when regime flag changes or you notice a clear regime shift.

Examples:
- "BTC ranging $77k-$80k since Monday, no clear breakout yet"
- "Alts pumping on Asia hours (00:00-04:00 UTC), quiet during US"
- "Funding broadly negative across majors → bullish bias for longs"

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

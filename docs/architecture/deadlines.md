# Deadlines and urgency

Applications may contain multiple deadlines. `DatePrecision` is `exact`, `day`, `month`, `approximate`, or `unknown`; `RawDeadlineText` preserves wording from a source. Exact/day precision requires `DeadlineAt`. Month, approximate, and unknown deadlines are not assigned invented timestamps.

Urgency values are `past_due`, `today`, `tomorrow`, `within_3_days`, `within_7_days`, `within_14_days`, `within_30_days`, `later`, or `unknown`. Vague deadlines always return `unknown`. The upcoming endpoint includes only exact/day deadlines and is a query foundation for later notifications—no scheduler or notification transport exists yet.

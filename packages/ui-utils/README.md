# Shared UI utilities

`@repo/ui-utils` owns browser-safe presentation helpers for BO and web. No API
clients, business calculations, environment reads or server configuration belong
here. Import explicit entrypoints; the date utility has no React dependency.

## Date formatting

```ts
import { formatDate, toDate } from "@repo/ui-utils/date";

formatDate("2026-10-24T08:00:00Z");
// In an Asia/Jakarta browser: "24 Okt 2026, 15.00"

formatDate("2026-10-24T08:00:00Z", {
	locale: "en-GB",
	timeZone: "Asia/Jakarta",
	format: { dateStyle: "long", timeStyle: "short" },
});
```

Defaults: `id-ID`, numeric day/year, abbreviated month, 24-hour hours/minutes,
and the runtime's local timezone. Set `locale` explicitly for another language.
`format` replaces the default Intl options (it is not merged), so date-only,
time-only and Intl style presets work without conflicting fields. `timeZone`
overrides the runtime zone, including any zone supplied inside `format`.
Locale punctuation and spacing follow the runtime's Intl/CLDR implementation.

Inputs are timezone-qualified ISO timestamps, `Date`, or epoch **milliseconds**.
Go fractional seconds and explicit UTC offsets are supported. `toDate` returns a
copy or `null`. Empty/invalid dates, date-only strings and timezone-less timestamps
return the configurable `fallback` (default `—`) rather than guessing a zone.
This utility formats instants, not calendar-only values such as birthdays.
Invalid formatter configuration throws rather than silently displaying wrong data.

## React and hydration

```tsx
import { useDateFormatter } from "@repo/ui-utils/react";

function Timestamp({ value }: { value: string }) {
	const formatDate = useDateFormatter();
	return <>{formatDate(value)}</>;
}
```

Use this adapter instead of calling `formatDate` directly during SSR. SSR and the
first hydration render return the same fallback; after hydration the browser
formats locally. No server timezone, environment reads, global mutable config,
hydration-warning suppression or client preference cookie is involved.

For semantic markup, prefer `DateTime` from `@repo/ui/date-time`. React components
belong in `packages/ui`, not this utility package. It uses the same adapter and
preserves the valid instant in the `<time dateTime>` attribute.

```tsx
import { DateTime } from "@repo/ui/date-time";

<DateTime value="2026-10-24T08:00:00Z" />
<DateTime value="2026-10-24T08:00:00Z" format={{ dateStyle: "long" }} />
```

Checks: `bun --no-env-file run --cwd packages/ui-utils check-types` and
`bun --no-env-file run --cwd packages/ui-utils test`. These focused tests cover
presentation logic only, with no services, database or network.

## Minimum pending presentation

`useMinimumPending(pending, minimumMs = 900)` from `@repo/ui-utils/react` keeps
a loading presentation visible for a minimum duration from the start of each
pending period. Long operations incur no extra hold after completion. Timers are
cleaned up on unmount or a new pending period. This is presentation only: callers
must handle errors and authorization redirects before applying the held state.

"use client";

import type { ComponentProps } from "react";
import { toDate, type DateInput, type DateFormatOptions } from "@repo/ui-utils/date";
import { useDateFormatter } from "@repo/ui-utils/react";

export type DateTimeProps = Omit<ComponentProps<"time">, "dateTime" | "children"> & DateFormatOptions & { value: DateInput };

/** Browser-local display, preserving the original instant as machine-readable UTC. */
export function DateTime({ value, locale, timeZone, format, fallback, ...props }: DateTimeProps) {
	const formatDate = useDateFormatter({ locale, timeZone, format, fallback });
	const date = toDate(value);
	return (
		<time {...props} dateTime={date?.toISOString()}>
			{formatDate(value)}
		</time>
	);
}

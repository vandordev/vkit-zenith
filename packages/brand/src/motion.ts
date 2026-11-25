// Motion JS mirror of the --motion-* / --ease-* CSS tokens (seconds). Keep in sync with tokens.css.
export const motion = {
	fast: 0.12,
	base: 0.2,
	slow: 0.32,
	easeOut: [0.22, 1, 0.36, 1],
	easeInOut: [0.65, 0, 0.35, 1],
} as const;

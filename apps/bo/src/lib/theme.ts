import { useEffect, useState } from "react";

export const themePreferences = ["light", "dark", "system"] as const;
export type ThemePreference = (typeof themePreferences)[number];
const KEY = "bo-theme";

// Pure per-browser convenience: storage may be blocked, so every access is guarded.
function read(): ThemePreference {
	try {
		return themePreferences.find((p) => p === localStorage.getItem(KEY)) ?? "system";
	} catch {
		return "system";
	}
}

export function applyTheme(preference: ThemePreference) {
	const dark = preference === "dark" || (preference === "system" && matchMedia("(prefers-color-scheme: dark)").matches);
	document.documentElement.classList.toggle("dark", dark);
	document.documentElement.style.colorScheme = dark ? "dark" : "light";
}

export function useTheme() {
	const [theme, setThemeState] = useState<ThemePreference>("system");
	useEffect(() => setThemeState(read()), []);
	useEffect(() => {
		applyTheme(theme);
		if (theme !== "system") return;
		const query = matchMedia("(prefers-color-scheme: dark)");
		const onChange = () => applyTheme("system");
		query.addEventListener("change", onChange);
		return () => query.removeEventListener("change", onChange);
	}, [theme]);
	const setTheme = (next: ThemePreference) => {
		setThemeState(next);
		try {
			localStorage.setItem(KEY, next);
		} catch {
			// Preference still applies for this page view.
		}
	};
	return [theme, setTheme] as const;
}

// Runs in <head> before first paint; mirrors read() + applyTheme() (keep in sync).
export const themeInitScript = `(function(){try{var p=null;try{p=localStorage.getItem("${KEY}")}catch(e){}var d=p==="dark"||(p!=="light"&&matchMedia("(prefers-color-scheme: dark)").matches);var r=document.documentElement;r.classList.toggle("dark",d);r.style.colorScheme=d?"dark":"light"}catch(e){}})()`;

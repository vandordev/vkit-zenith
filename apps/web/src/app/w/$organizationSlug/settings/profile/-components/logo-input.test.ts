import { expect, test } from "bun:test";
import { logoBody, logoMaxBytes, logoProblem } from "./logo-input";

test("png, jpeg and svg up to 1 MiB are accepted as metadata only", () => {
	expect(logoBody({ contentType: "image/png", size: 2048 })).toEqual({ contentType: "image/png", size: 2048 });
	expect(logoBody({ contentType: "image/svg+xml", size: logoMaxBytes })).toEqual({ contentType: "image/svg+xml", size: logoMaxBytes });
});

test("other formats, empty files and anything over 1 MiB are refused before any request", () => {
	for (const choice of [
		{ contentType: "image/gif", size: 10 },
		{ contentType: "application/pdf", size: 10 },
		{ contentType: "", size: 10 },
		{ contentType: "image/png", size: 0 },
		{ contentType: "image/jpeg", size: logoMaxBytes + 1 },
	]) {
		expect(logoProblem(choice)).toBeDefined();
		expect(logoBody(choice)).toBeUndefined();
	}
});

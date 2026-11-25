import { defineConfig } from "@hey-api/openapi-ts";
export default defineConfig({
	input: "../../contracts/openapi/web-api.json",
	output: "src/lib/api/generated",
	plugins: [
		"@hey-api/typescript",
		{ name: "@hey-api/client-fetch", baseUrl: false },
		{ name: "@hey-api/sdk", client: false },
		"@tanstack/react-query",
		{ name: "zod", compatibilityVersion: 4, dates: { offset: true }, requests: true, responses: true, definitions: true },
	],
});

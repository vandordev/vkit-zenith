import { defineConfig } from "@hey-api/openapi-ts";
export default defineConfig({
	input: "../../contracts/openapi/bo-api.json",
	output: "src/lib/api/generated",
	plugins: [
		"@hey-api/typescript",
		{ name: "@hey-api/client-fetch", baseUrl: false },
		{ name: "@hey-api/sdk", client: false },
		"@tanstack/react-query",
		{ name: "zod", compatibilityVersion: 4, requests: true, responses: true, definitions: true },
	],
});

import { expect, test } from "bun:test";
import { webToolEnvironment } from "./tooling";
test("tool adapter emits only listener settings", () => {
	expect(webToolEnvironment({ app: { environment: "development", logLevel: "info" }, web: { host: "127.0.0.1", port: 4200 } })).toEqual({ PORT: "4200", HOSTNAME: "127.0.0.1", NODE_ENV: "development" });
});

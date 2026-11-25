import { config as baseConfig } from "@repo/eslint-config/base";

export default [...baseConfig, { ignores: [".next/**", "out/**", "next-env.d.ts"] }];

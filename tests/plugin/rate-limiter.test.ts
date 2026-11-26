import { equal } from "node:assert/strict";
import { createFixedWindowLimiter } from "../../plugins/sdk-typescript/src/index.js";

const allow = createFixedWindowLimiter(2, 1000);
equal(allow("a", 0), true);
equal(allow("a", 1), true);
equal(allow("a", 999), false);
equal(allow("b", 999), true);
equal(allow("a", 1000), true);
equal(allow("a", 1001), true);
equal(allow("a", 1002), false);
equal(createFixedWindowLimiter(2, 1000)("a", 1002), true);

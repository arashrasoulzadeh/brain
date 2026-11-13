import { validateIdentifier, validateProjectId } from "../../plugins/sdk-typescript/src/index.js";

if (validateProjectId("dev-project") !== "dev-project") throw new Error("valid project id rejected");
try { validateProjectId("bad/project"); throw new Error("invalid project id accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project id") throw error; }
if (validateIdentifier("player_scores") !== "player_scores") throw new Error("valid identifier rejected");
try { validateIdentifier("player-scores"); throw new Error("invalid identifier accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid identifier") throw error; }

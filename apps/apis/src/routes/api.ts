import { Hono } from "hono";
import { authMiddleware } from "../middleware/auth";

export const apiRoutes = new Hono();

apiRoutes.use("*", authMiddleware);

apiRoutes.get("/me", (c) => {
  const user = c.get("user");
  return c.json({ user });
});

import { app } from "./app";

const port = Number(process.env.PORT) || 3001;

console.log(`Server is running on port ${port}`);

export default {
  port,
  fetch: app.fetch,
};

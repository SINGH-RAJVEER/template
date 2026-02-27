import { type Component, Show } from "solid-js";
import { authClient } from "./lib/auth-client";

const App: Component = () => {
  const session = authClient.useSession();

  return (
    <div style={{ padding: "2rem" }}>
      <h1>Template App</h1>
      <Show
        when={session()?.data?.user}
        fallback={
          <div>
            <p>You are not signed in.</p>
            <a href="/sign-in">Sign In</a> | <a href="/sign-up">Sign Up</a>
          </div>
        }
      >
        <div>
          <p>Welcome, {session()?.data?.user?.name}!</p>
          <button
            type="button"
            onClick={() => authClient.signOut()}
          >
            Sign Out
          </button>
        </div>
      </Show>
    </div>
  );
};

export default App;

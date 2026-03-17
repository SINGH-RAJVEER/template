import { authClient } from "./lib/auth-client";

const App = () => {
  const { data: session } = authClient.useSession();

  return (
    <div style={{ padding: "2rem" }}>
      <h1>Template App</h1>
      {session?.user ? (
        <div>
          <p>Welcome, {session.user.name}!</p>
          <button type="button" onClick={() => authClient.signOut()}>
            Sign Out
          </button>
        </div>
      ) : (
        <div>
          <p>You are not signed in.</p>
          <a href="/sign-in">Sign In</a> | <a href="/sign-up">Sign Up</a>
        </div>
      )}
    </div>
  );
};

export default App;

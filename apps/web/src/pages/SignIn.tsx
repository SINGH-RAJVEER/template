import { type Component, createSignal } from "solid-js";
import { authClient } from "../lib/auth-client";

export const SignIn: Component = () => {
  const [email, setEmail] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [error, setError] = createSignal<string | null>(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setError(null);

    const { error: authError } = await authClient.signIn.email({
      email: email(),
      password: password(),
      callbackURL: "/",
    });

    if (authError) {
      setError(authError.message ?? "Sign in failed");
    }
  };

  return (
    <div style={{ "max-width": "400px", margin: "4rem auto", padding: "2rem" }}>
      <h2>Sign In</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ "margin-bottom": "1rem" }}>
          <label for="email">Email</label>
          <input
            id="email"
            type="email"
            value={email()}
            onInput={(e) => setEmail(e.currentTarget.value)}
            required
            style={{ display: "block", width: "100%", padding: "0.5rem", "margin-top": "0.25rem" }}
          />
        </div>
        <div style={{ "margin-bottom": "1rem" }}>
          <label for="password">Password</label>
          <input
            id="password"
            type="password"
            value={password()}
            onInput={(e) => setPassword(e.currentTarget.value)}
            required
            style={{ display: "block", width: "100%", padding: "0.5rem", "margin-top": "0.25rem" }}
          />
        </div>
        {error() && <p style={{ color: "red" }}>{error()}</p>}
        <button type="submit" style={{ padding: "0.5rem 1rem" }}>
          Sign In
        </button>
        <p>
          Don't have an account? <a href="/sign-up">Sign Up</a>
        </p>
      </form>
    </div>
  );
};

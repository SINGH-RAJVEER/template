import { type Component, createSignal } from "solid-js";
import { authClient } from "../lib/auth-client";

export const SignUp: Component = () => {
  const [name, setName] = createSignal("");
  const [email, setEmail] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [error, setError] = createSignal<string | null>(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setError(null);

    const { error: authError } = await authClient.signUp.email({
      name: name(),
      email: email(),
      password: password(),
      callbackURL: "/",
    });

    if (authError) {
      setError(authError.message ?? "Sign up failed");
    }
  };

  return (
    <div style={{ "max-width": "400px", margin: "4rem auto", padding: "2rem" }}>
      <h2>Sign Up</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ "margin-bottom": "1rem" }}>
          <label for="name">Name</label>
          <input
            id="name"
            type="text"
            value={name()}
            onInput={(e) => setName(e.currentTarget.value)}
            required
            style={{ display: "block", width: "100%", padding: "0.5rem", "margin-top": "0.25rem" }}
          />
        </div>
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
          Sign Up
        </button>
        <p>
          Already have an account? <a href="/sign-in">Sign In</a>
        </p>
      </form>
    </div>
  );
};

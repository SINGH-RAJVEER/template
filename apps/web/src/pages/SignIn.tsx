import { Button } from "@template/ui/components/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@template/ui/components/card";
import { Input } from "@template/ui/components/input";
import { Label } from "@template/ui/components/label";
import { type FormEvent, useState } from "react";
import { authClient } from "../lib/auth-client";

export const SignIn = () => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError(null);

        const { error: authError } = await authClient.signIn.email({
            email,
            password,
            callbackURL: "/",
        });

        if (authError) setError(authError.message ?? "Sign in failed");
    };

    return (
        <main className="grid min-h-screen place-items-center px-4 py-12">
            <Card className="w-full max-w-md shadow-xl shadow-foreground/5">
                <form onSubmit={handleSubmit}>
                    <CardHeader>
                        <CardTitle className="text-2xl">Welcome back</CardTitle>
                        <CardDescription>Sign in with your email and password.</CardDescription>
                    </CardHeader>
                    <CardContent className="mt-6 space-y-5">
                        <div className="space-y-2">
                            <Label htmlFor="email">Email</Label>
                            <Input
                                id="email"
                                type="email"
                                autoComplete="email"
                                value={email}
                                onChange={(event) => setEmail(event.currentTarget.value)}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password">Password</Label>
                            <Input
                                id="password"
                                type="password"
                                autoComplete="current-password"
                                value={password}
                                onChange={(event) => setPassword(event.currentTarget.value)}
                                required
                            />
                        </div>
                        {error && <p className="text-sm text-destructive">{error}</p>}
                    </CardContent>
                    <CardFooter className="mt-6 flex-col items-stretch gap-4">
                        <Button type="submit" className="w-full">
                            Sign in
                        </Button>
                        <p className="text-center text-sm text-muted-foreground">
                            New here?{" "}
                            <a className="font-medium text-foreground underline" href="/sign-up">
                                Create an account
                            </a>
                        </p>
                    </CardFooter>
                </form>
            </Card>
        </main>
    );
};

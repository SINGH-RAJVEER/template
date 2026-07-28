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

export const SignUp = () => {
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError(null);

        const { error: authError } = await authClient.signUp.email({
            name,
            email,
            password,
            callbackURL: "/",
        });

        if (authError) {
            setError(authError.message ?? "Sign up failed");
        }
    };

    return (
        <main className="grid min-h-screen place-items-center px-4 py-12">
            <Card className="w-full max-w-md shadow-xl shadow-foreground/5">
                <form onSubmit={handleSubmit}>
                    <CardHeader>
                        <CardTitle className="text-2xl">Create your account</CardTitle>
                        <CardDescription>
                            Start with a name, email, and secure password.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="mt-6 space-y-5">
                        <div className="space-y-2">
                            <Label htmlFor="name">Name</Label>
                            <Input
                                id="name"
                                type="text"
                                autoComplete="name"
                                value={name}
                                onChange={(event) => setName(event.currentTarget.value)}
                                required
                            />
                        </div>
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
                                minLength={8}
                                maxLength={72}
                                autoComplete="new-password"
                                value={password}
                                onChange={(event) => setPassword(event.currentTarget.value)}
                                required
                            />
                        </div>
                        {error && <p className="text-sm text-destructive">{error}</p>}
                    </CardContent>
                    <CardFooter className="mt-6 flex-col items-stretch gap-4">
                        <Button type="submit" className="w-full">
                            Create account
                        </Button>
                        <p className="text-center text-sm text-muted-foreground">
                            Already registered?{" "}
                            <a className="font-medium text-foreground underline" href="/sign-in">
                                Sign in
                            </a>
                        </p>
                    </CardFooter>
                </form>
            </Card>
        </main>
    );
};

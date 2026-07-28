import { Button } from "@template/ui/components/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@template/ui/components/card";
import { authClient } from "./lib/auth-client";

const App = () => {
    const { data: session } = authClient.useSession();

    return (
        <main className="relative grid min-h-screen place-items-center overflow-hidden px-4 py-12">
            <div className="absolute -top-40 -right-24 size-96 rounded-full bg-primary/10 blur-3xl" />
            <div className="absolute -bottom-44 -left-24 size-96 rounded-full bg-accent/50 blur-3xl" />
            <Card className="relative w-full max-w-lg border-border/70 shadow-xl shadow-foreground/5">
                <CardHeader>
                    <p className="font-mono text-xs tracking-[0.22em] text-primary uppercase">
                        Go + React
                    </p>
                    <CardTitle className="text-3xl tracking-tight">Template App</CardTitle>
                    <CardDescription>
                        A typed starting point with a Go API and shared shadcn components.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {session?.user ? (
                        <div className="rounded-lg border bg-muted/45 p-4">
                            <p className="text-sm text-muted-foreground">Signed in as</p>
                            <p className="mt-1 text-lg font-medium">{session.user.name}</p>
                            <p className="text-sm text-muted-foreground">{session.user.email}</p>
                        </div>
                    ) : (
                        <p className="text-sm leading-6 text-muted-foreground">
                            Create an account or sign in to verify the complete authentication flow.
                        </p>
                    )}
                </CardContent>
                <CardFooter className="gap-3">
                    {session?.user ? (
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => authClient.signOut()}
                        >
                            Sign out
                        </Button>
                    ) : (
                        <>
                            <Button asChild>
                                <a href="/sign-up">Create account</a>
                            </Button>
                            <Button asChild variant="outline">
                                <a href="/sign-in">Sign in</a>
                            </Button>
                        </>
                    )}
                </CardFooter>
            </Card>
        </main>
    );
};

export default App;

export interface User {
    id: string;
    name: string;
    email: string;
    emailVerified: boolean;
    image?: string | null;
    createdAt: string;
    updatedAt: string;
}

export interface Session {
    id: string;
    userId: string;
    expiresAt: string;
    ipAddress?: string | null;
    userAgent?: string | null;
    createdAt: string;
    updatedAt: string;
}

export interface AuthSession {
    user: User;
    session: Session;
}

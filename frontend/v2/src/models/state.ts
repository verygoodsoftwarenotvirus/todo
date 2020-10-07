export class AuthStatus {
    isAuthenticated: boolean;
    isAdmin: boolean;

    constructor() {
      this.isAuthenticated = false;
      this.isAdmin = false;
    }
}

export interface LoginRequest {
    username: string;
    password: string;
    totpToken: string;
}

export interface RegistrationRequest {
    username: string;
    password: string;
    repeatedPassword: string;
}

export interface UserRegistrationResponse {
    id: number;
    username: string;
    isAdmin: boolean;
    qrCode: string;
    createdOn: number;
    lastUpdatedOn: number;
    archivedOn: number;
    passwordLastChangedOn: number;
}

export interface ErrorResponse {
    message: string;
    code: number;
}

export interface UserState {
    authStatus: AuthStatus;
}

export interface AppState {
    user: UserState;
}

// TODO: do these need to exist, and if so, do they need to exist here?

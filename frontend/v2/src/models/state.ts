export class LoginRequest {
    username: string;
    password: string;
    totpToken: string;

    constructor() {
        this.username = '';
        this.password = '';
        this.totpToken = '';
    }
}

export class RegistrationRequest {
    username: string;
    password: string;
    repeatedPassword: string;

    constructor() {
        this.username = '';
        this.password = '';
        this.repeatedPassword = '';
    }
}

export class TOTPTokenValidationRequest {
    userID: number;
    totpToken: string;

    constructor() {
        this.userID = 0;
        this.totpToken = '';
    }
}

// TODO: do these need to exist, and if so, do they need to exist here?

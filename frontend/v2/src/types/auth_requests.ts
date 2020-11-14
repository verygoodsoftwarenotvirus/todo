export class LoginRequest {
  username: string;
  password: string;
  totpToken: string;

  constructor() {
    this.username = "";
    this.password = "";
    this.totpToken = "";
  }
}

export class RegistrationRequest {
  username: string;
  password: string;
  repeatedPassword: string;

  constructor(
    username: string = "",
    password: string = "",
    repeatedPassword: string = ""
  ) {
    this.username = username;
    this.password = password;
    this.repeatedPassword = repeatedPassword;
  }
}

export class TOTPTokenValidationRequest {
  userID: number;
  totpToken: string;

  constructor(userID: number = 0, totpToken: string = "") {
    this.userID = userID;
    this.totpToken = totpToken;
  }
}

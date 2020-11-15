export type homePageTranslations = {
  mainGreeting: string;
  subGreeting: string;
};

export type registrationPageTranslations = {
  buttons: {
    register: string;
    submitVerification: string;
  };
  inputLabels: {
    username: string;
    password: string;
    passwordRepeat: string;
    twoFactorCode: string;
  };
  inputPlaceholders: {
    username: string;
    password: string;
    passwordRepeat: string;
  };
  linkTexts: {
    loginInstead: string;
  };
  notices: {
    saveQRSecretNotice: string;
  };
  instructions: {
    enterGeneratedTwoFactorCode: string;
  };
};

export type loginPageTranslations = {
  buttons: {
    login: string;
  };
  inputLabels: {
    username: string;
    password: string;
    twoFactorCode: string;
  };
  inputPlaceholders: {
    username: string;
    password: string;
    twoFactorCode: string;
  };
  linkTexts: {
    forgotPassword: string;
    createAccount: string;
  };
};

export type userSettingsPageTranslations = {
  myAccount: string;
  buttons: {
    updateUserInfo: string;
    changePassword: string;
  };
  sectionLabels: {
    userInfo: string;
    password: string;
  };
  inputLabels: {
    username: string;
    emailAddress: string;
    currentPassword: string;
    newPassword: string;
    twoFactorToken: string;
  };
  inputPlaceholders: {
    currentPassword: string;
    newPassword: string;
    twoFactorToken: string;
  };
};

export type siteSettingsPageTranslations = {
  title: string;
  buttons: {
    cycleCookieSecret: string;
  };
  confirmations: {
    cycleCookieSecret: string;
  };
  sectionLabels: {
    actions: string;
  };
};

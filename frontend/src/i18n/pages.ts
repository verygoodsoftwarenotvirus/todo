import type { webhookModelTranslations } from '@/i18n/types';

export type homePageTranslations = {
  mainGreeting: string;
  subGreeting: string;
  navBar: {
    serviceName: string;
    buttons: {
      login: string;
      register: string;
    };
  };
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
    twoFactorCode: string;
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

export type userAdminPageTranslations = {
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

export type accountSettingsPageTranslations = {
  title: string;
  sectionLabels: {
    info: string;
    members: string;
  };
  buttons: {
    saveMembers: string;
  };
  inputLabels: {
    name: string;
    members: string;
  };
  inputPlaceholders: {
    name: string;
  };
};

export type userSettingsPageTranslations = {
  title: string;
  buttons: {
    updateUserInfo: string;
    changePassword: string;
  };
  sectionLabels: {
    userInfo: string;
    password: string;
  };
  valueLabels: {
    reputation: string;
  };
  hovertexts: {
    reputation: string;
  };
  inputLabels: {
    username: string;
    emailAddress: string;
    currentPassword: string;
    newPassword: string;
    twoFactorToken: string;
  };
  inputPlaceholders: {
    email: string;
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

export type webhookCreationPageTranslations = {
  model: webhookModelTranslations;
  validInputs: {
    events: string[];
    types: string[];
  };
};

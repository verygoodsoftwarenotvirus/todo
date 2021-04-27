export type apiTableTranslations = {
  page: string;
  delete: string;
  perPage: string;
  inputPlaceholders: {
    search: string;
  };
};
export type auditLogEntryTableTranslations = {
  title: string;
  page: string;
  perPage: string;
  inputPlaceholders: {
    search: string;
  };
  columns: {
    id: string;
    eventType: string;
    context: string;
    createdOn: string;
  };
};

export type userDropdownTranslations = {
  settings: string;
  adminMode: string;
  logout: string;
};

export type adminNavbarTranslations = {
  dashboard: string;
};

export type authNavbarTranslations = {
  serviceName: string;
};

export type homepageNavbarTranslations = {
  serviceName: string;
  buttons: {
    login: string;
    register: string;
  };
};

export type primarySidebarTranslations = {
  serviceName: string;
  things: string;
  admin: string;
  settings: string;
  userSettings: string;
  accountSettings: string;
  serverSettings: string;
  accounts: string;
  users: string;
  webhooks: string;
  apiClients: string;
  auditLog: string;
  items: string;
};

export type mainFooterTranslations = {
  keepInTouch: string;
  weLikeYou: string;
  usefulLinks: string;
  aboutUs: string;
  blog: string;
  otherResources: string;
  termsAndConditions: string;
  privacyPolicy: string;
  contactUs: string;
};

export type adminFooterTranslations = {
  copyright: string;
  aboutUs: string;
  blog: string;
};

export type smallFooterTranslations = {
  copyright: string;
  aboutUs: string;
  blog: string;
};

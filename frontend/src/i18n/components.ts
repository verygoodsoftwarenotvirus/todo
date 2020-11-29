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
  };
};

export type primarySidebarTranslations = {
  serviceName: string;
  things: string;
  admin: string;
  users: string;
  oauth2Clients: string;
  webhooks: string;
  auditLog: string;
  serverSettings: string;
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

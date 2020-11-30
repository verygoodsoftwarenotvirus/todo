export type itemModelTranslations = {
  actions: {
    create: string;
  };
  columns: {
    id: string;
    name: string;
    details: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToUser: string;
  };
  labels: {
    name: string;
    details: string;
  };
  inputPlaceholders: {
    name: string;
    details: string;
  };
};

export type auditLogEntryTranslations = {
  columns: {
    id: string;
    eventType: string;
    context: string;
    createdOn: string;
  };
};

export type userModelTranslations = {
  actions: {
    save: string;
    ban: string;
  };
  columns: {
    id: string;
    username: string;
    isAdmin: string;
    requiresPasswordChange: string;
    passwordLastChangedOn: string;
    createdOn: string;
    lastUpdatedOn: string;
    archivedOn: string;
  };
  labels: {
    name: string;
  };
  inputPlaceholders: {
    name: string;
  };
};

export type oauth2ClientModelTranslations = {
  actions: {
    create: string;
    update: string;
  };
  columns: {
    id: string;
    name: string;
    clientID: string;
    clientSecret: string;
    redirectURI: string;
    scopes: string;
    implicitAllowed: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToUser: string;
  };
  labels: {
    name: string;
    clientID: string;
    clientSecret: string;
    redirectURI: string;
  };
  inputPlaceholders: {
    name: string;
    redirectURI: string;
  };
};

export type webhookModelTranslations = {
  actions: {
    create: string;
    update: string;
  };
  columns: {
    id: string;
    name: string;
    contentType: string;
    url: string;
    method: string;
    events: string;
    dataTypes: string;
    topics: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToUser: string;
  };
  labels: {
    name: string;
    contentType: string;
    url: string;
    method: string;
    events: string;
    dataTypes: string;
    topics: string;
  };
  inputPlaceholders: {
    name: string;
    contentType: string;
    url: string;
    method: string;
  };
};

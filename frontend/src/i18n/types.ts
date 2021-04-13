export type auditLogEntryTranslations = {
  columns: {
    id: string;
    eventType: string;
    context: string;
    createdOn: string;
  };
};

export type accountUserMembershipModelTranslations = {
  actions: {
    create: string;
  };
  columns: {
    id: string;
    createdOn: string;
    archivedOn: string
    userAccountPermissions: string;
    defaultAccount: string;
    belongsToAccount: string;
    belongsToUser: string;
  };
  labels: {
    name: string;
    accountSubscriptionPlanID: string;
  };
  inputPlaceholders: {
    name: string;
    accountSubscriptionPlanID: string;
  };
};

export type accountModelTranslations = {
  actions: {
    create: string;
  };
  columns: {
    id: string;
    externalID: string;
    name: string;
    accountSubscriptionPlanID: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToUser: string;
    defaultNewMemberPermissions: string;
    archivedOn: string;
  };
  labels: {
    name: string;
    accountSubscriptionPlanID: string;
  };
  inputPlaceholders: {
    name: string;
    accountSubscriptionPlanID: string;
  };
};

export type userModelTranslations = {
  actions: {
    save: string;
    ban: string;
  };
  columns: {
    id: string;
    externalID: string;
    username: string;
    reputation: string;
    reputationExplanation: string;
    serviceAdminPermissions: string;
    requiresPasswordChange: string;
    passwordLastChangedOn: string;
    createdOn: string;
    lastUpdatedOn: string;
    archivedOn: string;
  };
  labels: {
    id: string;
    username: string;
    isAdmin: string;
    requiresPasswordChange: string;
    passwordLastChangedOn: string;
    createdOn: string;
    lastUpdatedOn: string;
    archivedOn: string;
  };
  inputPlaceholders: {
    username: string;
  };
};

export type webhookModelTranslations = {
  actions: {
    create: string;
    update: string;
  };
  columns: {
    id: string;
    externalID: string;
    name: string;
    contentType: string;
    url: string;
    method: string;
    events: string;
    dataTypes: string;
    topics: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToAccount: string;
    archivedOn: string;
  };
  labels: {
    name: string;
    contentType: string;
    url: string;
    method: string;
    events: string;
    dataTypes: string;
    topics: string;
    createdOn: string;
  };
  inputPlaceholders: {
    name: string;
    contentType: string;
    url: string;
    method: string;
  };
};

export type apiClientModelTranslations = {
  actions: {
    create: string;
  };
  columns: {
    id: string;
    externalID: string;
    name: string;
    clientID: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToAccount: string;
    archivedOn: string;
  };
  labels: {
    name: string;
    clientID: string;
  };
  inputPlaceholders: {
    name: string;
    clientID: string;
  };
};

export type itemModelTranslations = {
  actions: {
    create: string;
  };
  columns: {
    id: string;
    externalID: string;
    name: string;
    details: string;
    createdOn: string;
    lastUpdatedOn: string;
    archivedOn: string;
    belongsToAccount: string;
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

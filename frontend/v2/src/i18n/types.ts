export type itemModelTranslations = {
  actions: {
    create: string;
  },
  columns: {
    id: string;
    name: string;
    details: string;
    createdOn: string;
    lastUpdatedOn: string;
    belongsToUser: string;
  },
  labels: {
    name: string;
    details: string;
  },
  inputPlaceholders: {
    name: string;
    details: string;
  },
}

export type userModelTranslations = {
  actions: {
    save: string;
    delete: string;
  },
  labels: {
    name: string;
  },
  inputPlaceholders: {
    name: string;
  },
}

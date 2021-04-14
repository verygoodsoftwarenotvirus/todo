import type { SiteTranslationMap } from './definitions';
import type { webhookModelTranslations } from './types';

const _id = 'ID',
  _externalID = 'External ID',
  _createdOn = 'Created On',
  _archivedOn = 'Archived On',
  _login = 'Login',
  _name = 'Name',
  _username = 'Username',
  _password = 'Password',
  _lastUpdatedOn = 'Last Updated On',
  _belongsToUser = 'Belongs to User',
  _belongsToAccount = 'Belongs to Account',
  _serviceName = 'Todo',
  _settings = 'Settings',
  _copyright = 'Copyright ©',
  _aboutUs = 'About Us';

const webhook: webhookModelTranslations = {
  actions: {
    create: 'Create',
    update: 'Update',
  },
  columns: {
    id: _id,
    externalID: _externalID,
    name: _name,
    contentType: 'Content-Type',
    url: 'URL',
    method: 'Method',
    events: 'Events',
    dataTypes: 'Data Types',
    topics: 'Topics',
    createdOn: _createdOn,
    lastUpdatedOn: _lastUpdatedOn,
    belongsToAccount: _belongsToAccount,
    archivedOn: _archivedOn,
  },
  labels: {
    name: _name,
    contentType: 'Content-Type',
    url: 'URL',
    method: 'Methods',
    events: 'Events',
    dataTypes: 'Data Types',
    topics: 'Topics',
    createdOn: _createdOn,
  },
  inputPlaceholders: {
    name: _name,
    contentType: 'application/example',
    url: 'https://url-to-use.com',
    method: 'POST',
  },
};

export const englishTranslations: SiteTranslationMap = {
  components: {
    apiTable: {
      page: 'Page',
      delete: 'Delete',
      perPage: 'per page',
      inputPlaceholders: {
        search: 'Search...',
      },
    },
    auditLogEntryTable: {
      title: 'Audit Log Entries',
      page: 'page',
      perPage: 'per page',
      inputPlaceholders: {
        search: 'search',
      },
      columns: {
        id: _id,
        eventType: 'Event Type',
        context: 'Context',
        createdOn: _createdOn,
      },
    },
    dropdowns: {
      userDropdown: {
        settings: _settings,
        adminMode: 'Admin Mode',
        logout: 'Log Out',
      },
    },
    navbars: {
      adminNavbar: {
        dashboard: 'Dashboard',
      },
      authNavbar: {
        serviceName: _serviceName,
      },
    },
    sidebars: {
      primary: {
        serviceName: _serviceName,
        items: 'Items',
        things: 'Things',
        admin: 'Admin',
        users: 'Users',
        apiClients: 'API Clients',
        webhooks: 'Webhooks',
        accounts: 'Accounts',
        auditLog: 'Audit Log',
        settings: _settings,
        accountSettings: 'Account',
        userSettings: 'User',
        serverSettings: 'Server',
      },
    },
    footers: {
      mainFooter: {
        keepInTouch: "Let's keep in touch!",
        weLikeYou: 'We like you.',
        usefulLinks: 'Useful Links',
        aboutUs: _aboutUs,
        blog: 'Blog',
        otherResources: 'Other Resources',
        termsAndConditions: 'Terms & Conditions',
        privacyPolicy: 'Privacy Policy',
        contactUs: 'Contact Us',
      },
      adminFooter: {
        copyright: _copyright,
        aboutUs: _aboutUs,
        blog: 'Blog',
      },
      smallFooter: {
        copyright: _copyright,
        aboutUs: _aboutUs,
        blog: 'Blog',
      },
    },
  },
  pages: {
    home: {
      mainGreeting: 'this is the homepage.',
      subGreeting: 'websites are cool and good to read.',
      navBar: {
        serviceName: _serviceName,
        buttons: {
          login: _login,
          register: 'Register',
        },
      },
    },
    login: {
      buttons: {
        login: _login,
      },
      inputLabels: {
        username: _username,
        password: _password,
        twoFactorCode: '2FA Code',
      },
      inputPlaceholders: {
        username: _username.toLowerCase(),
        password: 'pick something strong, please',
        twoFactorCode: '123456',
      },
      linkTexts: {
        forgotPassword: 'Forgot your authentication?',
        createAccount: 'Create account',
      },
    },
    registration: {
      buttons: {
        register: 'Create Account',
        submitVerification: "I've Saved It!",
      },
      inputLabels: {
        username: _username,
        password: _password,
        passwordRepeat: 'Confirm Password',
        twoFactorCode: '2FA Code',
      },
      inputPlaceholders: {
        username: _username.toLowerCase(),
        password: 'your authentication',
        passwordRepeat: 'your authentication again',
        twoFactorCode: '123456',
      },
      linkTexts: {
        loginInstead: 'Login instead?',
      },
      notices: {
        saveQRSecretNotice:
          "Save the secret this QR code contains in your 2FA Code generator of choice. You'll be required to generate a token from it on every login.",
      },
      instructions: {
        enterGeneratedTwoFactorCode:
          "Enter an example generated code to verify you've completed the above step:",
      },
    },
    accountSettings: {
      title: 'Account settings',
      sectionLabels: {
        info: 'Info',
        members: 'Members',
      },
      buttons: {
        saveMembers: 'Save Members',
      },
      inputLabels: {
        name: 'Name',
        members: 'Members',
      },
      inputPlaceholders: {
        name: 'name',
      },
    },
    userSettings: {
      title: 'User settings',
      buttons: {
        updateUserInfo: 'Update',
        changePassword: 'Change Password',
      },
      sectionLabels: {
        userInfo: 'User Info',
        password: _password,
      },
      inputLabels: {
        username: _username,
        emailAddress: 'Email Address',
        currentPassword: 'Current Password',
        newPassword: 'New Password',
        twoFactorToken: '2FA Token',
      },
      valueLabels: {
        reputation: 'Account Status',
      },
      hovertexts: {
        reputation: 'account status: ',
      },
      inputPlaceholders: {
        email: "we don't want your stinkin' email!",
        currentPassword: 'current authentication',
        newPassword: 'new authentication',
        twoFactorToken: '123456',
      },
    },
    siteSettings: {
      title: 'Site settings',
      buttons: {
        cycleCookieSecret: 'Cycle Cookie Secret',
      },
      confirmations: {
        cycleCookieSecret:
          'Are you sure you want to cycle the cookie secret? This will effectively log out every user.',
      },
      sectionLabels: {
        actions: 'Actions',
      },
    },
    webhookCreationPage: {
      model: webhook,
      validInputs: {
        events: ['All', 'Create', 'Update', 'Delete'],
        types: ['All', 'Item'],
        topics: ['All'],
      },
    },
    userAdminPage: {
      myAccount: 'undefined FUCK YOU undefined',
      buttons: {
        updateUserInfo: 'undefined FUCK YOU undefined',
        changePassword: 'undefined FUCK YOU undefined',
      },
      sectionLabels: {
        userInfo: 'undefined FUCK YOU undefined',
        password: 'undefined FUCK YOU undefined',
      },
      inputLabels: {
        username: 'undefined FUCK YOU undefined',
        emailAddress: 'undefined FUCK YOU undefined',
        currentPassword: 'undefined FUCK YOU undefined',
        newPassword: 'undefined FUCK YOU undefined',
        twoFactorToken: 'undefined FUCK YOU undefined',
      },
      inputPlaceholders: {
        currentPassword: 'undefined FUCK YOU undefined',
        newPassword: 'undefined FUCK YOU undefined',
        twoFactorToken: 'undefined FUCK YOU undefined',
      },
    },
  },
  models: {
    user: {
      actions: {
        save: 'Save',
        ban: 'Ban',
      },
      columns: {
        id: 'ID',
        externalID: 'External ID',
        username: _username,
        reputation: 'Reputation',
        reputationExplanation: 'Reputation Explanation',
        serviceAdminPermissions: 'Service Admin Permissions',
        requiresPasswordChange: 'Requires Password Change',
        passwordLastChangedOn: 'Password Last Changed On',
        createdOn: 'Created On',
        lastUpdatedOn: 'Last Updated On',
        archivedOn: _archivedOn,
      },
      labels: {
        id: 'ID',
        username: _username,
        isAdmin: 'Is Admin',
        requiresPasswordChange: 'Requires Password Change?',
        passwordLastChangedOn: 'Password Last Changed',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        archivedOn: _archivedOn,
      },
      inputPlaceholders: {
        username: 'new username',
      },
    },
    auditLogEntry: {
      columns: {
        id: _id,
        eventType: 'Event Type',
        context: 'Context',
        createdOn: _createdOn,
      },
    },
    webhook: webhook,
    apiClient: {
      actions: {
        create: 'Create Account',
      },
      columns: {
        id: _id,
        externalID: _externalID,
        name: _name,
        clientID: 'Client ID',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        belongsToUser: _belongsToUser,
        archivedOn: _archivedOn,
      },
      labels: {
        name: 'Client Name',
        username: _username,
        password: _password,
        totpToken: 'TOTP Token',
      },
      inputPlaceholders: {
        name: 'Client Name',
        username: _username,
        password: _password,
        totpToken: 'TOTP Token',
      },
    },
    account: {
      actions: {
        create: 'Create Account',
      },
      columns: {
        id: _id,
        externalID: _externalID,
        name: _name,
        accountSubscriptionPlanID: 'Plan ID',
        createdOn: _createdOn,
        defaultNewMemberPermissions: 'Default New User Permissions',
        lastUpdatedOn: _lastUpdatedOn,
        belongsToUser: _belongsToUser,
        archivedOn: _archivedOn,
      },
      labels: {
        name: _name,
        accountSubscriptionPlanID: 'Plan ID',
      },
      inputPlaceholders: {
        name: 'name',
        accountSubscriptionPlanID: 'Plan ID',
      },
    },
    accountUserMembership: {
      actions: {
        create: 'Create Account',
      },
      columns: {
        id: _id,
        createdOn: _createdOn,
        belongsToUser: "User ID",
        userAccountPermissions: "Permissions",
        defaultAccount: "Default",
        belongsToAccount: _belongsToAccount,
        archivedOn: _archivedOn,
      },
      labels: {
        name: _name,
        accountSubscriptionPlanID: 'Plan ID',
      },
      inputPlaceholders: {
        name: 'name',
        accountSubscriptionPlanID: 'Plan ID',
      },
    },
    item: {
      actions: {
        create: 'Create Item',
      },
      columns: {
        id: _id,
        externalID: _externalID,
        name: _name,
        details: 'Details',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        belongsToAccount: _belongsToAccount,
        archivedOn: _archivedOn,
      },
      labels: {
        name: _name,
        details: 'Details',
      },
      inputPlaceholders: {
        name: 'name',
        details: 'details',
      },
    },
  },
};

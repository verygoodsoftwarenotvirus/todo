import type { SiteTranslationMap } from '@/i18n/definitions';

const _id = 'ID',
  _createdOn = 'Created On',
  _name = 'Name',
  _lastUpdatedOn = 'Last Updated On',
  _belongsToUser = 'Belongs to User';

export const englishTranslations: SiteTranslationMap = {
  components: {
    apiTable: {
      page: 'Page',
      delete: 'Delete',
      perPage: 'per page:',
      inputPlaceholders: {
        search: 'Search...',
      },
    },
    auditLogEntryTable: {
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
        settings: 'Settings',
        adminMode: 'Admin Mode',
        logout: 'Log Out',
      },
    },
    navbars: {
      adminNavbar: {
        dashboard: 'Dashboard',
      },
      authNavbar: {
        serviceName: 'Todo',
      },
      homepageNavbar: {
        serviceName: 'Todo',
        buttons: {
          login: 'Login',
        },
      },
    },
    sidebars: {
      primary: {
        serviceName: 'Todo',
        things: 'Things',
        admin: 'Admin',
        users: 'Users',
        oauth2Clients: 'OAuth2 Clients',
        webhooks: 'Webhooks',
        auditLog: 'Audit Log',
        serverSettings: 'Server Settings',
        items: 'Items',
      },
    },
    footers: {
      mainFooter: {
        keepInTouch: "Let's keep in touch!",
        weLikeYou: 'We like you.',
        usefulLinks: 'Useful Links',
        aboutUs: 'About Us',
        blog: 'Blog',
        otherResources: 'Other Resources',
        termsAndConditions: 'Terms & Conditions',
        privacyPolicy: 'Privacy Policy',
        contactUs: 'Contact Us',
      },
      adminFooter: {
        copyright: 'Copyright ©',
        aboutUs: 'About Us',
        blog: 'Blog',
      },
      smallFooter: {
        copyright: 'Copyright ©',
        aboutUs: 'About Us',
        blog: 'Blog',
      },
    },
  },
  pages: {
    home: {
      mainGreeting: 'this is the homepage.',
      subGreeting: 'websites are cool and good to read.',
    },
    login: {
      buttons: {
        login: 'Login',
      },
      inputLabels: {
        username: 'Username',
        password: 'Password',
        twoFactorCode: '2FA Code',
      },
      inputPlaceholders: {
        username: 'username',
        password: 'pick something strong, please',
        twoFactorCode: '123456',
      },
      linkTexts: {
        forgotPassword: 'Forgot your password?',
        createAccount: 'Create account',
      },
    },
    registration: {
      buttons: {
        register: 'Create Account',
        submitVerification: "I've Saved It!",
      },
      inputLabels: {
        username: 'Username',
        password: 'Password',
        passwordRepeat: 'Confirm Password',
        twoFactorCode: '2FA Code',
      },
      inputPlaceholders: {
        username: 'username',
        password: 'your password',
        passwordRepeat: 'your password again',
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
    userSettings: {
      myAccount: 'My account',
      buttons: {
        updateUserInfo: 'Update',
        changePassword: 'Change Password',
      },
      sectionLabels: {
        userInfo: 'User Info',
        password: 'Password',
      },
      inputLabels: {
        username: 'Username',
        emailAddress: 'Email Address',
        currentPassword: 'Current Password',
        newPassword: 'New Password',
        twoFactorToken: '2FA Token',
      },
      inputPlaceholders: {
        currentPassword: 'current password',
        newPassword: 'new password',
        twoFactorToken: '123456',
      },
    },
  },
  models: {
    user: {
      actions: {
        save: 'Save',
        delete: 'Delete',
      },
      columns: {
        id: 'ID',
        username: 'Username',
        isAdmin: 'Admin',
        requiresPasswordChange: 'Requires Password Change',
        passwordLastChangedOn: 'Password Last Changed On',
        createdOn: 'Created On',
        lastUpdatedOn: 'Last Updated On',
        archivedOn: 'Archived On',
      },
      labels: {
        name: _name,
      },
      inputPlaceholders: {
        name: 'name',
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
    oauth2Client: {
      actions: {
        create: 'Create',
        update: 'Update',
      },
      columns: {
        id: _id,
        name: _name,
        clientID: 'Client ID',
        clientSecret: 'Client Secret',
        redirectURI: 'Redirect URI',
        scopes: 'Scopes',
        implicitAllowed: 'Implicit Allowed',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        belongsToUser: _belongsToUser,
      },
      labels: {
        name: _name,
        clientID: 'Client ID',
        clientSecret: 'Client Secret',
        redirectURI: 'Redirect URI',
      },
      inputPlaceholders: {
        name: _name,
        redirectURI: 'https://redirect-to-here.pizza',
      },
    },
    webhook: {
      actions: {
        create: 'Create',
        update: 'Update',
      },
      columns: {
        id: _id,
        name: _name,
        contentType: 'Content-Type',
        url: 'URL',
        method: 'Method',
        events: 'Events',
        dataTypes: 'Data Types',
        topics: 'Topics',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        belongsToUser: _belongsToUser,
      },
      labels: {
        name: _name,
        contentType: 'Content-Type',
        url: 'URL',
        method: 'Methods',
        events: 'Events',
        dataTypes: 'Data Types',
        topics: 'Topics',
      },
      inputPlaceholders: {
        name: _name,
        contentType: 'application/example',
        url: 'https://url-to-use.com',
        method: 'POST',
      },
    },
    item: {
      actions: {
        create: 'Create Item',
      },
      columns: {
        id: _id,
        name: _name,
        details: 'Details',
        createdOn: _createdOn,
        lastUpdatedOn: _lastUpdatedOn,
        belongsToUser: _belongsToUser,
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

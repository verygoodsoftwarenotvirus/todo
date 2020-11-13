import {
    saveItem,
    fetchItem,
    createItem,
    deleteItem,
    searchForItems,
    fetchListOfItems, fetchAuditLogEntriesForItem,
} from './items';

import {
    login,
    logout,
    selfRequest,
    registrationRequest,
    passwordChangeRequest,
    checkAuthStatusRequest,
    validateTOTPSecretWithToken,
    twoFactorSecretChangeRequest,
} from './auth';

import {
    saveUser,
    fetchUser,
    deleteUser,
    fetchListOfUsers,
} from './users';

export class V1APIClient {
    // users stuff
    static fetchUser = fetchUser;
    static fetchListOfUsers = fetchListOfUsers;
    static saveUser = saveUser;
    static deleteUser = deleteUser;

    // auth stuff
    static login = login;
    static logout = logout;
    static selfRequest = selfRequest;
    static passwordChangeRequest = passwordChangeRequest;
    static twoFactorSecretChangeRequest = twoFactorSecretChangeRequest;
    static registrationRequest = registrationRequest;
    static checkAuthStatusRequest = checkAuthStatusRequest;
    static validateTOTPSecretWithToken = validateTOTPSecretWithToken;

    // items stuff
    static createItem = createItem;
    static fetchItem = fetchItem;
    static saveItem = saveItem;
    static deleteItem = deleteItem;
    static searchForItems = searchForItems;
    static fetchListOfItems = fetchListOfItems;
    static fetchAuditLogEntriesForItem = fetchAuditLogEntriesForItem;
}
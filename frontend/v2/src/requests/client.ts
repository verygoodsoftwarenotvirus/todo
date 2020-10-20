import {
    fetchListOfItems,
    createItem,
    saveItem,
    fetchItem,
    deleteItem,
    searchForItems,
} from './items';

import {
    selfRequest,
    checkAuthStatusRequest,
    validateTOTPSecretWithToken,
    registrationRequest,
    login,
    logout,
} from './auth';

import {
    fetchUser,
    fetchListOfUsers,
    saveUser,
    deleteUser,
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
}
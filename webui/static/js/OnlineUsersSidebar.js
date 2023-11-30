//import { use } from "chai";
import { STRINGS } from "./ConstantStrings.js";

export default class OnlineUsersSidebar {
    constructor(DOMElements, switchChildView, webSocketManager, chatView) {
        this.DOMElements = DOMElements;
        this.switchChildView = switchChildView;
        this.webSocketManager = webSocketManager;
        this.chatView = chatView;
        
        this.webSocketManager.on("onlineUsers", this.handleOnlineUsersListReceived);
        this.webSocketManager.on("newOnlineUser", this.handleNewOnlineUserReceived);
        this.webSocketManager.on("offlineUser", this.handleOfflineUserReceived);

        this.currentSessionUserData = {};
        this.currentChatRecipientData = {};
        this.onlineUsersArray = [];
    }

    initalize() {
        this.DOMElements.onlineUsersList.addEventListener("click", this.handleOnlineUserClick);
        this.adjustSidebarPosition();
    }

    uninitalize() {
        this.DOMElements.onlineUsersList.removeEventListener("click", this.handleOnlineUserClick);
    }

    // ---------------------------- ON INITALIZATION --------------------------

    //To keep the sidebar on position when scrolling
    adjustSidebarPosition = () => {
        const navbarHeight = this.DOMElements.navBar.offsetHeight;
        const onlineUsersWrapper = document.getElementById('onlineUsersWrapper');
        onlineUsersWrapper.style.top = navbarHeight + 40 + 'px';
    }

    // --------- HANDLE ONLINE USERS LIST AFTER LOGIN OR PAGE RELOAD ----------

    //Handles online users message received from server
    handleOnlineUsersListReceived = async (payload) => {
        if (payload.result !== STRINGS.SUCCESS) {
            console.log("Error: Could not get online users");
            return
        }

        this.clearOnlineUsersList();

        //As online users data will include also current user, filter it out
        const filteredOnlineUsers = await this.filterOnlineUsers(payload.data);

        //If there is online users, add to DOM and array. If not, show notice to user
        if (filteredOnlineUsers.length !== 0) {
            this.updateOnlineUsersList(filteredOnlineUsers);
        } else {
            this.showNoOnlineUsersText();
        }
    }

    //Filter out the current user and return new online users list
    filterOnlineUsers = async (usersData) => {
        if (usersData === null || !(await this.confirmCurrentUsernameSet())) {
            return [];
        }
        return usersData.filter(userData => userData.name !== this.currentSessionUserData.username);
    }

    //Add users to DOM and js array for sorting, return boolean if there was online users
    updateOnlineUsersList = (onlineUsers) => {
        onlineUsers.forEach((userData) => {
            this.addOnlineUserToEndOfUsersList(userData);
            this.addOnlineUserToArray(userData);
        });
    }

    clearOnlineUsersList = () => {
        this.DOMElements.onlineUsersList.innerHTML = "";
    }

    // ------------------------ CHANGE IN USERS LIST --------------------------

    handleNewOnlineUserReceived = (payload) => {

        this.addOnlineUserToArray(payload.data);

        //Sorts array by last message time and for users without any messages, alphabetically
        this.sortOnlineUsersArray();

        //If there was no online users previously, remove no online users notice
        if (this.onlineUsersArray.length === 1) {
            this.removeNoOnlineUsersText();
        }

        //Get userID who is next in list after new user for reference in adding to DOM
        const nextUserID = this.getNextUserIDinList(String(payload.data.id));

        //Add to DOM to appropriate position
        this.addOnlineUserBeforeOtherUser(payload.data, nextUserID);

        //If online user is currect chat recipient, make their name unclickable and inform ChatView
        if (payload.data.name === this.currentChatRecipientData.username) {
            this.makeCurrentChatUsernameUnclickable();
            this.chatView.handleChatRecipientOnline();
        }
    };

    handleOfflineUserReceived = (payload) => {

        const userID = String(payload.data.id);

        this.removeUserFromDOM(userID);
        this.removeUserFromArray(userID);

        //If it was last online user, show no-one online text
        if (this.onlineUsersArray.length === 0) {
            this.showNoOnlineUsersText();
        }

        //If offline user was current chat recipient, notify Chat View
        if (payload.data.name === this.currentChatRecipientData.username) {
            this.chatView.handleChatRecipientGoneOffline();
        }
    };

    // ---------- UPDATING USERS ORDER ON INCOMING/OUTGOING MESSAGE -----------

    handleListAndArrayOnNewMessage = (userID) => {

        //Change last message time to now on recipint in users array
        //(Users array won't be sorted here, due to no usage.)
        this.changeLastMessageTimeToNowByUserID(userID);

        //Move the element in DOM to upmost position
        this.moveUserToTheTop(userID);
    }

    // ------------------------- JS ONLINE USERS ARRAY ------------------------

    addOnlineUserToArray = (userData) => {

        //If last message date exists, convert it to milliseconds. Otherwise leave lastMessageTime undefined.
        let lastMessageTime;
        if (userData.lastMessageDate !== "") {
            lastMessageTime = new Date(userData.lastMessageDate).getTime()
        }
        
        this.onlineUsersArray.push({
            id: String(userData.id),
            username: userData.name,
            lastMessageTime: lastMessageTime
        });
    }

    //Sorts array by last message time and for users without any messages, alphabetically
    sortOnlineUsersArray = () => {
        this.onlineUsersArray.sort((a, b) => {
            if (a.lastMessageTime && b.lastMessageTime) { //If both have any last message time set
                return b.lastMessageTime - a.lastMessageTime
            }
            if (a.lastMessageTime) { //If only a has some message sent, they should be before b
                return -1
            }
            if (b.lastMessageTime) { //If only b has some message sent, they should be before a
                return 1
            }
            return a.username.localeCompare(b.username) //If neither has any messages, sort alphabetically
        });
    }

    getNextUserIDinList = (newUserID) => {
        const index = this.onlineUsersArray.findIndex(user => user.id === newUserID);

        if (index === this.onlineUsersArray.length - 1) { //If new user is last one in sorted list return empty string
            return ""
        }

        return this.onlineUsersArray[index + 1].id
    }

    //For removing user from array when they go offline
    removeUserFromArray = (offlineUserID) => {
        this.onlineUsersArray = this.onlineUsersArray.filter(userData => userData.id !== offlineUserID);
    }

    changeLastMessageTimeToNowByUserID = (userID) => {
        const userObject = this.onlineUsersArray.find(user => user.userID === userID)

        if (userObject) {
            userObject.lastMessageTime = new Date().getTime();
        }
    }

    // -------------------- ADD/REMOVE/MOVE USER IN DOM -----------------------

    generateOnlineUserHTML = (userData) => {
        return `
        <div class="onlineUser selectableUser ps-2 pb-1 d-inline-flex align-items-center" data-id=${userData.id} data-username=${userData.name}>
            <div class="greenCircle"></div>
            <span>${userData.name}</span>
            <div class="newMsgNotificationBubble">0</div>
        </div>`
    }

    addOnlineUserToEndOfUsersList = (userData) => {
        const userHTML = this.generateOnlineUserHTML(userData);
        this.DOMElements.onlineUsersList.insertAdjacentHTML("beforeend", userHTML);
    }

    //For adding user to DOM before some other user (if no other, then to the end of list)
    addOnlineUserBeforeOtherUser = (userData, nextUserID) => {
        const userHTML = this.generateOnlineUserHTML(userData);
        if (nextUserID === "") { //If there is no next user, add to the end of list
            this.DOMElements.onlineUsersList.insertAdjacentHTML("beforeend", userHTML);
        } else {
            const nextUserEl = this.DOMElements.onlineUsersList.querySelector(`[data-id="${nextUserID}"]`);
            nextUserEl.insertAdjacentHTML("beforebegin", userHTML);
        }
    }

    removeUserFromDOM = (userID) => {
        const userEl = this.DOMElements.onlineUsersList.querySelector(`[data-id="${userID}"]`);
        if (userEl) {
            userEl.remove();
        } else {
            console.log("Error: Couldn't remove offline user from DOM");
        }
    }

    moveUserToTheTop = (userID) => {
        const userEl = this.DOMElements.onlineUsersList.querySelector(`[data-id="${userID}"]`);
        if (userEl && this.DOMElements.onlineUsersList.firstChild !== userEl) {
            this.DOMElements.onlineUsersList.prepend(userEl);
        }
    }

    // -------------------------- CLICKING ON USERS ---------------------------

    handleOnlineUserClick = (event) => {
        //Get the user element closest to click
        const target = event.target.closest('.onlineUser');

        if (target) {
            const recipientUserID = target.dataset.id;
            const recipientUsername = target.dataset.username;

            //Do nothing if chat with this user is already open
            if (recipientUsername === this.currentChatRecipientData.username) {
                return
            }

            //Send open chat request if user doesn't have this chat already opened
            this.webSocketManager.sendOpenChatRequest(Number(recipientUserID));
        }
    }

    makePreviousChatUsernameClickable = () => {
        if (this.currentChatRecipientData.username === "") {
            return
        }

        const usernameEl = document.querySelector(`.onlineUser[data-username="${this.currentChatRecipientData.username}"]`);

        if (usernameEl !== null) {
            usernameEl.classList.add("selectableUser");
        }
    }

    makeCurrentChatUsernameUnclickable = () => {
        if (!this.currentChatRecipientData.username) {
            return
        }

        const usernameEl = document.querySelector(`.onlineUser[data-username="${this.currentChatRecipientData.username}"]`);

        if (usernameEl !== null) {
            usernameEl.classList.remove("selectableUser");
        }
    }

    // -------------------------- HANDLE OPENING CHAT -------------------------

    handleOpeningChat = (recipientUserID, recipientUsername) => {
        this.resetNotificationsBubble(recipientUsername);
        this.makePreviousChatUsernameClickable();
        this.storeCurrentChatRecipientData(recipientUserID, recipientUsername);
        this.makeCurrentChatUsernameUnclickable();
    }

    // ------------------------- NOTIFICATION BUBBLES -------------------------

    showAndIncreaseNotificationsBubble = (username) => {
        const usernameEl = document.querySelector(`.onlineUser[data-username="${username}"]`);
        const notificationBubbleEl = usernameEl.querySelector(".newMsgNotificationBubble");
        const currentNotificationsAmount = Number(notificationBubbleEl.textContent);
        notificationBubbleEl.textContent = currentNotificationsAmount + 1;
        notificationBubbleEl.style.display = "flex";
    }

    resetNotificationsBubble = (username) => {
        const usernameEl = document.querySelector(`.onlineUser[data-username="${username}"]`);
        const notificationBubbleEl = usernameEl.querySelector(".newMsgNotificationBubble");
        notificationBubbleEl.textContent = "0";
        notificationBubbleEl.style.display = "none";
    }

    // -------------- STORING CURRENT USER AND CHAT RECIPIENT DATA ------------

    updateCurrentUserData = (userData) => {
        this.currentSessionUserData = userData;
    }

    storeCurrentChatRecipientData = (userID, username) => {
        this.currentChatRecipientData = {
            userID: userID,
            username: username
        }
    }

    removeCurrentChatRecipientData = () => {
        this.currentChatRecipientData = {};
    }

    // ---------------------------- HELPER METHODS ----------------------------

    //After login, wait once the current session username is set, before updating the online users list (current user will be also included there)
    confirmCurrentUsernameSet = async () => {
        const MAX_ATTEMPTS = 10;
        const DELAY = 100;
        let attempts = 0;

        while (!this.currentSessionUserData.username && attempts < MAX_ATTEMPTS) {
            await new Promise(res => setTimeout(res, DELAY));
            attempts++;
        }

        if (!this.currentSessionUserData.username) {
            console.error("Current session username not set after maximum attempts.");
            return false
        }

        return true
    }

    //To notify user if no other users are online
    showNoOnlineUsersText = () => {
        const noUsersHTML = `<div id="nooneOnlineText">No users online</div>`
        this.DOMElements.onlineUsersList.insertAdjacentHTML("afterbegin", noUsersHTML);
    }

    removeNoOnlineUsersText = () => {
        const noUsersEl = document.getElementById("nooneOnlineText");
        if (noUsersEl) {
            noUsersEl.remove();
        }
    }
}
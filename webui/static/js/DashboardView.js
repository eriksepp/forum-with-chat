import PostsListView from "./PostsListView.js";
import FullPostView from "./FullPostView.js";
import ChatView from "./ChatView.js";
import OnlineUsersSidebar from "./OnlineUsersSidebar.js";
import { getCurrentISODate } from "./helpers.js";
import { STRINGS } from "./ConstantStrings.js";

export default class DashboardView {
    constructor(DOMElements, switchView, webSocketManager) {
        this.DOMElements = DOMElements;
        this.switchView = switchView;
        this.webSocketManager = webSocketManager;

        //Add listeners for incoming ws messages
        this.webSocketManager.on("fullPostAndCommentsReply", this.handleFullPostReply);
        this.webSocketManager.on("newPostReply", this.handleNewPostReply);
        this.webSocketManager.on("inputChatMessage", this.handleIncomingChatMessage);
        this.webSocketManager.on("sendMessageToOpendChatReply", this.handleOutgoingChatMessageResponse);
        this.webSocketManager.on("openChatReply", this.handleOpenChatReply);

        //Initalize child views
        this.childViews = {};
        this.childViews.postsList = new PostsListView(this.DOMElements, this.switchChildView, this.webSocketManager);
        this.childViews.fullPost = new FullPostView(this.DOMElements, this.switchChildView, this.webSocketManager);
        this.childViews.chat = new ChatView(this.DOMElements, this.switchChildView, this.webSocketManager);
        this.childViews.onlineUsers = new OnlineUsersSidebar(this.DOMElements, this.switchChildView, this.webSocketManager, this.childViews.chat);

        //State variables
        this.viewStack = [];
        this.currentChildView = "";
        this.currentSessionUserData = {};
        this.currentChatRecipientData = {};
    }

    show() {
        this.DOMElements.navBar.style.display = "block";
        this.DOMElements.dashboardContainer.style.display = "block";
        this.bindEventListeners();
        this.childViews.onlineUsers.initalize();
        this.childViews.postsList.emptyPostsListAndGetTenNewestPosts();
        this.switchChildView(STRINGS.POSTS_LIST);
    }

    hide() {
        this.DOMElements.navBar.style.display = "none";
        this.DOMElements.dashboardContainer.style.display = "none";
        this.childViews.onlineUsers.uninitalize();
        this.unbindEventListeners();
    }

    // ---------------------------- EVENT LISTENERS ---------------------------

    bindEventListeners = () => {
        this.DOMElements.forumTitle.addEventListener("click", this.updateAndNavitagateToPostsList);
        this.DOMElements.logOutButton.addEventListener("click", this.handleLogOutClick);
        this.DOMElements.createNewPostModal.addEventListener("hidden.bs.modal", this.clearNewPostForm);
        this.DOMElements.createNewPostForm.addEventListener("submit", this.handleSubmittingNewPost);
        Array.from(this.DOMElements.arrowBackButtons).forEach((arrow) => arrow.addEventListener("click", this.switchToPreviousView));
    }

    unbindEventListeners = () => {
        this.DOMElements.forumTitle.removeEventListener("click", this.updateAndNavitagateToPostsList);
        this.DOMElements.logOutButton.removeEventListener("click", this.handleLogOutClick);
        this.DOMElements.createNewPostModal.removeEventListener("hidden.bs.modal", this.clearNewPostForm);
        this.DOMElements.createNewPostForm.removeEventListener("submit", this.handleSubmittingNewPost);
        Array.from(this.DOMElements.arrowBackButtons).forEach((arrow) => arrow.removeEventListener("click", this.switchToPreviousView));
    }

    // ------------------------ INITALIZING USER SESSION ----------------------

    //Store users ID and username to this and other views
    storeCurrentSessionUserData = (userData) => {
        this.currentSessionUserData = {
            userID: String(userData.id),
            username: userData.name
        }
        this.childViews.chat.updateCurrentUserData(this.currentSessionUserData);
        this.childViews.onlineUsers.updateCurrentUserData(this.currentSessionUserData);
    }

    showUsernameInNavbar = () => {
        document.getElementById("navbarUsername").textContent = "Welcome, " + this.currentSessionUserData.username + "!";
    }

    // ------------------------- HANDLING LOG OUT CLICK -----------------------

    handleLogOutClick = () => {
        this.webSocketManager.sendLogOutRequest();
    }

    // ----------------------------- VIEW CHANGING ----------------------------

    //Main function to switch between posts feed, full post and chat view
    switchChildView = (viewName) => {
        //If some view is already shown, hide it
        if (this.currentChildView) {
            this.childViews[this.currentChildView].hide();
        }

        //If new view is different from previous, store it to view stack
        if (this.viewStack[this.viewStack.length - 1] !== viewName) {
            this.viewStack.push(viewName);
        }
        
        //Update and show new view
        this.currentChildView = viewName;
        this.childViews[this.currentChildView].show();
    }

    //Used to switch to previous view in stack when user presses UI back arrow button
    switchToPreviousView = () => {
        //If view stack is empty, do nothing
        if (this.viewStack.length <= 1) {
            return
        }

        //If current view was chat then close the chat
        if (this.currentChildView === STRINGS.CHAT) {
            this.handleClosingChat();
        }

        //Remove last view from view stack
        this.viewStack.pop();

        //Switch to previous view in view stack
        const previousViewName = this.viewStack[this.viewStack.length - 1]
        this.switchChildView(previousViewName);
    }

    //After Forum title click, go to fo posts feed view and show 10 latest posts
    updateAndNavitagateToPostsList = () => {
        //If current view was chat then close the chat
        if (this.currentChildView === STRINGS.CHAT) {
            this.handleClosingChat();
        }

        //Empty view stack
        this.viewStack = [];

        //Get 10 latest posts and navigate to posts feed
        this.childViews.postsList.emptyPostsListAndGetTenNewestPosts();
        this.switchChildView(STRINGS.POSTS_LIST);
    }

    //When navigating away from chat send close chat message to server and reset chat related elements and variables
    handleClosingChat = () => {
        this.webSocketManager.sendCloseChatRequest();
        this.childViews.onlineUsers.makePreviousChatUsernameClickable();
        this.currentChatRecipientData = {};
        this.childViews.onlineUsers.removeCurrentChatRecipientData();
    }

    // --------------------------- NEW POST CREATION --------------------------

    //Runs when users presses "Create post" button
    handleSubmittingNewPost = (event) => {
        event.preventDefault();
        this.cleanNewPostFormErrors();

        //Get new post data from form
        const formData = new FormData(event.target);
        const postTheme = formData.get("title");
        const postText = formData.get("text");
        const postCategories = formData.getAll("categoriesID");

        //Convert categoryIDs from strings to integers
        postCategories.forEach((value, index, array) => {
            array[index] = Number(value);
        });

        //Check for errors in new post form
        const error = this.validateNewPostFormData(postTheme, postText, postCategories);

        //If any error exists, show message to user and stop this function
        if (error) {
          this.showNewPostCreationError(error.field, error.message);
          return;
        }

        //If new post form was valid, send new post request to server
        const date = getCurrentISODate();
        this.webSocketManager.sendCreateNewPostRequest(date, postTheme, postText, postCategories);
    }

    //Function to check if all new post form fields are correctly filled
    validateNewPostFormData = (postTheme, postText, postCategories) => {
        if (postCategories.length < 1) {
            return { field: "newPostCategoryWrapper", message: "Select at least 1 category for your post" };
        } else if (postTheme.trim() === "") {
            return { field: "newPostTitle", message: "Post title missing"};
        } else if (postText.trim() === "") { 
            return { field: "newPostText", message: "Post text missing"};
        }
    }

    //Function to show error for corresponding input field during new post creation
    showNewPostCreationError = (fieldID, message) => {
        document.getElementById(fieldID).classList.add("is-invalid");
        document.getElementById(`${fieldID}Label`).innerHTML = message;
    }

    //Resets new post form and removes any error messages
    clearNewPostForm = () => {
        this.DOMElements.createNewPostForm.reset();
        this.cleanNewPostFormErrors();
    }

    cleanNewPostFormErrors() {
        document.getElementById("newPostCategoryWrapper").classList.remove("is-invalid")
        document.getElementById("newPostCategoryLabel").innerHTML = "Pick a category";
        document.getElementById("newPostTitle").classList.remove("is-invalid")
        document.getElementById("newPostTitleLabel").innerHTML = "Post title";
        document.getElementById("newPostText").classList.remove("is-invalid")
        document.getElementById("newPostTextLabel").innerHTML = "Post text";
    }

    // ----------------------------- CHAT RELATED -----------------------------

    //Function to handle a reply for a request to open a chat
    handleOpenChatReply = (payload) => {
        if (payload.result !== STRINGS.SUCCESS) {
            console.log("Error: could not open chat")
            return
        }

        //Get chat recipient ID and username
        const recipientUserID = String(payload.data.recipientUser.id);
        const recipientUsername = payload.data.recipientUser.name;

        //Set online users list accordingly (remove notification bubble if exists, make new recipient unclickable and previous one if exists clickable)
        this.childViews.onlineUsers.handleOpeningChat(recipientUserID, recipientUsername);

        //Store new chat recipient ID and username
        this.storeChatRecipientUserData(recipientUserID, recipientUsername);

        //Parse previous 10 messages and store new hat recipient ID and username in chat view
        this.childViews.chat.initializeChat(recipientUserID, recipientUsername, payload.data.message);

        this.switchChildView(STRINGS.CHAT);
    }

    storeChatRecipientUserData = (userID, username) => {
        this.currentChatRecipientData = {
            userID: userID,
            username: username
        }
    }

    removeChatRecipientData = () => {
        this.currentChatRecipientData = {};
        this.childViews.onlineUsers.removeCurrentChatRecipientData();
    }

    handleIncomingChatMessage = (payload) => {
        const senderUsername = payload.data.author.name;

        //Inform Online Users Sidebar to reorder the users
        this.childViews.onlineUsers.handleListAndArrayOnNewMessage(payload.data.author.id);
    
        //If chat with the user who sent message is not yet opened, show notification bubble in online users list
        if (this.currentChatRecipientData.username !== senderUsername) {
            const senderUsername = payload.data.author.name;
            this.childViews.onlineUsers.showAndIncreaseNotificationsBubble(senderUsername);
        } else { //If chat is already opened, show new message in the chat view
            this.childViews.chat.handleReceivingNewMessage(payload);
        }
    }

    handleOutgoingChatMessageResponse = (payload) => {
        if (payload.result !== STRINGS.SUCCESS) {
            console.log("Error: Could not send message")
            return
        }

        this.childViews.onlineUsers.handleListAndArrayOnNewMessage(this.currentChatRecipientData.userID);
        this.childViews.chat.handleMessageSent();
    }

    // ------------------- HANDLING POST RELATED REPLIES ----------------------

    handleFullPostReply = (payload) => {
        this.childViews.fullPost.setFullPostData(payload);
        this.switchChildView(STRINGS.FULL_POST);
    }

    handleNewPostReply = (payload) => {
        if (payload.result !== STRINGS.SUCCESS) {
            console.log("Error: could not create new post")
            return
        }
        //Hide new post creation modal
        const bsNewPostModal = bootstrap.Modal.getInstance(this.DOMElements.createNewPostModal);
        bsNewPostModal.hide();

        this.childViews.postsList.resetPostsList();
        this.childViews.postsList.handleReceivedPostsPortion(payload);
        this.switchChildView(STRINGS.POSTS_LIST);
    }
}
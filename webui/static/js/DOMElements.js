export const DOMElements = {

    // MAIN WRAPPER FOR LOGIN AND REGISTER VIEW -------------------------------
    loginRegisterWrapper: document.getElementById("loginRegisterWrapper"),

    // LOGIN VIEW -------------------------------------------------------------
    loginContainer: document.getElementById("loginContainer"),
    loginWelcomeToText: document.getElementById("loginWelcomeToTxt"),
    loginTitle: document.getElementById("loginTitle"),
    loginSpinner: document.getElementById("loginSpinner"),
    loginForm: document.getElementById("loginForm"),
    loginUsernameInput: document.getElementById("loginUsername"),
    goToRegisterButton: document.getElementById("goToRegisterButton"),

    // REGISTER VIEW ----------------------------------------------------------
    registerContainer: document.getElementById("registerContainer"),
    registerTitle: document.getElementById("registerTitle"),
    registerForm: document.getElementById("registerForm"),
    passwordFieldForRegister: document.getElementById('registerPassword'),
    peekPasswordIcon: document.getElementById('togglePassword'),
    goToLoginButton: document.getElementById("goBackToLoginButton"),

    // DASHBOARD VIEW ---------------------------------------------------------
    dashboardContainer: document.getElementById("mainContainer"),
    navBar: document.getElementById("navbar"),
    forumTitle: document.getElementById("forumTitle"),
    logOutButton: document.getElementById("logout"),
    onlineUsersList: document.getElementById("onlineUsers"),
    arrowBackButtons: document.getElementsByClassName("backBtn"),
    
    // POSTS FEED VIEW --------------------------------------------------------
    postsListContainer: document.getElementById("postsList"),

    // FULL POST VIEW ---------------------------------------------------------
    fullPostContainer: document.getElementById("fullPostContainer"),
    fullPostTitle: document.getElementById("fullPostTitle"),
    fullPostUsername: document.getElementById("fullPostUsername"),
    fullPostCommentAmount: document.getElementById("fullPostCommentAmount"),
    fullPostCategoriesWrapper: document.getElementById("fullPostCategoriesWrapper"),
    fullPostContent: document.getElementById("fullPostContent"),
    fullPostIdForComment: document.getElementById("newCommentPostId"),
    newCommentForm: document.getElementById("createCommentForm"),
    newCommentInput: document.getElementById("newCommentText"),
    commentFormButtons: document.getElementById("commentFormButtons"),
    cancelNewComment: document.getElementById("cancelCreateComment"),
    fullPostCommentsContainer: document.getElementById("fullPostCommentsContainer"),

    // CHAT VIEW --------------------------------------------------------------
    chatContainer: document.getElementById("chat-container"),
    chatTitle: document.getElementById("chatTitle"),
    chatMessagesWrapper: document.getElementById("chatMessagesContainer"),
    chatMessages: document.getElementById("chatMessages"),
    chatMessageForm: document.getElementById("chatMessageForm"),
    chatInput: document.getElementById("chatMessageInput"),
    sendMessageButton: document.getElementById("sendMessageBtn"),
    userOfflineOverlay: document.getElementById("chatUserOfflineOverlay"),
    userOfflineMessage: document.getElementById("chatUserOfflineText"),
    zeroMessagesText: document.getElementById("noMessagesText"),

    // NEW POST CREATION MODAL ------------------------------------------------
    createNewPostModal: document.getElementById("createPostModal"),
    createNewPostForm: document.getElementById("createPostForm"),
}
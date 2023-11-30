import { throttleAndDebounce } from "./helpers.js";
import { STRINGS } from "./ConstantStrings.js";

export default class ChatView {
    constructor(DOMElements, switchChildView, webSocketManager) {
        this.DOMElements = DOMElements;
        this.switchChildView = switchChildView;
        this.webSocketManager = webSocketManager;
        this.throttleAndDebouncedHandleScroll = throttleAndDebounce(this.handleScroll.bind(this), 300);
        
        this.webSocketManager.on("chatPortionReply", this.handleChatPortionReply)

        this.topMessageID = 0;
        this.chatRecipientOnline = true;
        this.gettingMoreMessages = false;
        this.pendingSentMessageData = {};
        this.currentSessionUserData = {};
        this.currentChatRecipientData = {};
    }

    show() {
        this.DOMElements.chatContainer.style.display = "flex";
        this.resetChatInput();
        this.bindEventListeners();
        this.updateUserOfflineOverlayPosition();
        this.scrollToChatBottom();
        this.chatRecipientOnline = true;
        this.gettingMoreMessages = false;
    }

    hide() {
        this.DOMElements.chatContainer.style.display = "none";
        this.unbindEventListeners();
        window.removeEventListener('resize', this.updateUserOfflineOverlayPosition);
        this.DOMElements.userOfflineOverlay.style.display = "none";
    }

    // ---------------------------- EVENT LISTENERS ---------------------------

    bindEventListeners = () => {
        this.DOMElements.chatInput.addEventListener("input", this.activateSendBtnIfValidMsg);
        this.DOMElements.chatMessageForm.addEventListener("submit", this.handleNewMessageSubmit);
    }

    unbindEventListeners = () => {
        this.DOMElements.chatInput.removeEventListener("input", this.activateSendBtnIfValidMsg);
        this.DOMElements.chatMessageForm.removeEventListener("submit", this.handleNewMessageSubmit);
    }

    // --------------------------- INITALIZING CHAT ---------------------------

    //Runs when opening chat
    initializeChat = (recipientUserID, recipientUsername, messages) => {

        //Store recipient data
        this.setChatUsername(recipientUsername);
        this.storeCurrentChatRecipientData(recipientUserID, recipientUsername);

        //Clear previous messages
        this.DOMElements.chatMessages.innerHTML = ""

        //If no old messages, show notice to user
        if (messages === null || messages.length === 0) {
            this.DOMElements.zeroMessagesText.style.display = "block";
            return
        }

        //If there are old messages, remove "no old messages" notice
        this.DOMElements.zeroMessagesText.style.display = "none";

        this.parsePreviousMessagesData(messages);

        this.scrollToChatBottom();

        //If there may be more old message, listen for scroll
        if (messages.length >= 10) {
            this.addScrollListening();
        };
    }

    setChatUsername = (username) => {
        this.DOMElements.chatTitle.textContent = `Chat with ${username}`;
    }

    storeCurrentChatRecipientData = (recipientUserID, recipientUsername) => {
        this.currentChatRecipientData = {
            userID: recipientUserID,
            username: recipientUsername
        };
    }

    // ------------------------- SHOWING PAST MESSAGES ------------------------

    //To show old messages received from server
    parsePreviousMessagesData = (messages) => {
        for (let i = 0; i < messages.length; i++) {
            const message = messages[i]
            
            const { userClass, username } = this.determineUserClassAndUsername(message);

            if (userClass === null) {
                console.log("Invalid chat message received");
                return;
            }

            const messageDate = this.formatDateToLocalString(new Date(message.dateCreate));
            const messageContent = message.content;

            this.addMessageToDOM(userClass, username, messageDate, messageContent);

            //Store the top message ID for requesting more previous ones from server
            if (i === messages.length - 1) {
                this.topMessageID = messages[i].id;
            }
        }
    }

    generateMessageHTML = (userClass, username, messageDate, messageContent) => {
        return `<div class="message ${userClass}">
                    <div class="message-info">${username} - ${messageDate}</div>
                    <div class="message-bubble">
                        <span class="message-content">${messageContent}</span>
                    </div>
                </div>`
    }

    //For adding message to DOM. Possibility to choose if to add to the top or bottom of message list
    addMessageToDOM = (userClass, username, messageDate, messageContent, addToEnd = false) => {
        const messageHTML = this.generateMessageHTML(userClass, username, messageDate, messageContent);
        let direction;
        if (!addToEnd) {
            direction = "afterbegin";
        } else {
            direction = "beforeend";
        };
        this.DOMElements.chatMessages.insertAdjacentHTML(direction, messageHTML);
    }

    // --------------------------- HANDLE SCROLLING ---------------------------

    //If user scrolls high enough, request more old messages from server
    handleScroll = () => {
        const offset = this.DOMElements.chatMessages.clientHeight / 2;
        if (!this.gettingMoreMessages && this.DOMElements.chatMessages.scrollTop <= offset) {
            this.gettingMoreMessages = true;
            this.webSocketManager.sendChatPortionRequest(this.topMessageID);
        }
    }

    scrollToChatBottom = () => {
        this.DOMElements.chatMessagesWrapper.scrollTop = this.DOMElements.chatMessagesWrapper.scrollHeight;
        this.DOMElements.chatMessages.scrollTop = this.DOMElements.chatMessages.scrollHeight;
    }

    //Switch scroll listening on
    addScrollListening = () => {
        this.addLoadingSpinner();
        this.DOMElements.chatMessages.addEventListener("scroll", this.throttleAndDebouncedHandleScroll);
    }

    //If user has reached the end of old messages, stop listening for scroll
    removeScrollListening = () => {
        this.DOMElements.chatMessages.removeEventListener("scroll", this.throttleAndDebouncedHandleScroll);
    }

    //To avoid browser "jumping" to the top new element when adding old messages
    scrollToPreviousPositionAfterLoadingMoreMessages = (scrollTopFromBottom) => {
        let newScrollTop = this.DOMElements.chatMessages.scrollHeight - scrollTopFromBottom;
        this.DOMElements.chatMessages.scrollTo({ top: newScrollTop, behavior: "instant" });
    }

    // --------------------- GETTING MORE PAST MESSAGES -----------------------

    //When receiving more old messages from server
    handleChatPortionReply = (payload) => {
        setTimeout(() => {
            const currentScrollTopFromBottom = this.DOMElements.chatMessages.scrollHeight - this.DOMElements.chatMessages.scrollTop;

            this.removeLoadingSpinner();

            const messages = payload.data.message;

            if (messages === null) {
                this.handleNoMorePreviousMessages();
                return
            }

            this.parsePreviousMessagesData(messages);
    
            this.manageStatusForGettingMoreMessages(messages);
    
            this.scrollToPreviousPositionAfterLoadingMoreMessages(currentScrollTopFromBottom);
        }, 600)
    }

    //To stop listening for scroll when no more old messages or if there may be more, adding loading spinner
    manageStatusForGettingMoreMessages = (messages) => {
        if (messages.length < 10) {
            this.handleNoMorePreviousMessages();
        } else {
            this.gettingMoreMessages = false;
            this.addLoadingSpinner();
        }
    }

    //If there are no more old messages to request from server
    handleNoMorePreviousMessages = () => {
        this.removeScrollListening();
        this.addChatBeginningText();
    }

    //If there are no more old messages to request from server show notice to user
    addChatBeginningText = () => {
        const chatBeginningText = `
        <div id="chatBeginningText" class="d-flex justify-content-center mb-2">
            <span>You've reached the Beginning</span>
        </div>`
        this.DOMElements.chatMessages.insertAdjacentHTML("afterbegin", chatBeginningText);
    }

    // ---------------------------- LOADING SPINNER ---------------------------

    addLoadingSpinner = () => {
        const spinnerHTML = `
        <div id="chatLoadingSpinner" class="d-flex justify-content-center mb-2">
            <div class="spinner-border text-primary" role="status"></div>
        </div>`
        this.DOMElements.chatMessages.insertAdjacentHTML("afterbegin", spinnerHTML);
    }

    removeLoadingSpinner = () => {
        const loadingSpinner = document.getElementById("chatLoadingSpinner");
        loadingSpinner.classList.remove("d-flex");
        loadingSpinner.style.display = "none";
    }

    // --------------- RECIPIENT GOING OFFLINE / COMING ONLINE ----------------

    //When receiving info that current recipient is online
    handleChatRecipientOnline = () => {
        //Do nothing if chat recipient has been online since last online users update
        if (this.chatRecipientOnline) {
            return
        }

        //If they have been offline previously but came back, resume chat
        this.DOMElements.userOfflineOverlay.style.display = "none";
        window.removeEventListener('resize', this.updateUserOfflineOverlayPosition);
        this.chatRecipientOnline = true;
    }

    //If current recipient goes online, disable message sending, show notice to user
    handleChatRecipientGoneOffline = () => {
        this.DOMElements.userOfflineMessage.innerHTML = `${this.currentChatRecipientData.username} has gone offline<br>You can continue messaging when they are back`
        this.DOMElements.userOfflineOverlay.style.display = "block";
        this.updateUserOfflineOverlayPosition();
        window.addEventListener('resize', this.updateUserOfflineOverlayPosition);
        this.chatRecipientOnline = false;
        this.DOMElements.sendMessageButton.classList.remove("sendMessageBtnActive");
    }

    //To resize the "Recipient has gone offline" notice when user changes window size
    updateUserOfflineOverlayPosition = () => {
        const chatMessagesPosition = this.DOMElements.chatMessagesWrapper.getBoundingClientRect();
        this.DOMElements.userOfflineOverlay.style.top = chatMessagesPosition.top + "px";
        this.DOMElements.userOfflineOverlay.style.left = chatMessagesPosition.left + "px";
        this.DOMElements.userOfflineOverlay.style.width = chatMessagesPosition.width + "px";
        this.DOMElements.userOfflineOverlay.style.height = chatMessagesPosition.height + "px";
    }

    // -------------------------- SENDING A MESSAGE ---------------------------

    //For validating new message and sending it to server
    handleNewMessageSubmit = (event) => {
        event.preventDefault();

        if (this.validChatMessageWritten() && this.chatRecipientOnline) {
            const dateObj = new Date();
            const isoDate = dateObj.toISOString();
            const message = this.DOMElements.chatInput.value;
            this.pendingSentMessageData = { date: dateObj, messageContent: message };
            this.webSocketManager.sendChatMessage(isoDate, message);
        }
    }

    //If message is not empty, it's valid
    validChatMessageWritten = () => {
        if (this.DOMElements.chatInput.value.trim() !== "") {
            return true
        } 
        return false
    }

    //Enable and disable send button according to if user has written something to message input
    activateSendBtnIfValidMsg = (event) => {
        if (this.validChatMessageWritten() && this.chatRecipientOnline) {
            this.DOMElements.sendMessageButton.classList.add("sendMessageBtnActive");
        } else {
            this.DOMElements.sendMessageButton.classList.remove("sendMessageBtnActive");
        }
    }

    //When new message has been sent succesfuly, add it to messages list and clear input
    handleMessageSent = () => {
        this.DOMElements.chatInput.value = "";
        this.DOMElements.chatInput.placeholder = "";
        this.DOMElements.sendMessageButton.classList.remove("sendMessageBtnActive");
        this.DOMElements.zeroMessagesText.style.display = "none";

        const userClass = "currentUser";
        const username = "You";
        const messageDate = this.formatDateToLocalString(this.pendingSentMessageData.date);
        const messageContent = this.pendingSentMessageData.messageContent;
        this.addMessageToDOM(userClass, username, messageDate, messageContent, true);
        this.scrollToChatBottom();
    }

    //Clear message input field and disable send button
    resetChatInput = () => {
        this.DOMElements.chatInput.value = "";
        this.DOMElements.chatInput.placeholder = "Type message";
        this.DOMElements.sendMessageButton.classList.remove("sendMessageBtnActive");
    }

    // -------------------------- RECEIVING A MESSAGE -------------------------

    //When reciving new message to opened chat, show it in messages list
    handleReceivingNewMessage = (payload) => {
        this.DOMElements.zeroMessagesText.style.display = "none";
        const userClass = "otherUser";
        const username = payload.data.author.name;
        const messageDate = this.formatDateToLocalString(new Date(payload.data.date));
        const messageContent = payload.data.messageContent;
        this.addMessageToDOM(userClass, username, messageDate, messageContent, true);
        this.scrollToChatBottom();
    }

    // -------------------------------- HELPERS -------------------------------

    //For adding new message to message list with correct CSS class ("currentUser" or "otherUser")
    determineUserClassAndUsername = (message) => {
        let userClass;
        let username;
        if (String(message.author.id) === this.currentSessionUserData.userID) {
            userClass = "currentUser";
            username = "You";
        } else if (String(message.author.id) === this.currentChatRecipientData.userID) {
            userClass = "otherUser";
            username = message.author.name;
        } else {
            return { userClass: null, username: null }
        };

        return { userClass, username }
    }

    //To show date & time for message in correct format
    formatDateToLocalString = (dateData) => {
        const dateOptions = { day: '2-digit', month: '2-digit', year: '2-digit' };
        const formattedDate = dateData.toLocaleDateString("en-GB", dateOptions);

        const timeOptions = { hour: '2-digit', minute: '2-digit' };
        const formattedTime = dateData.toLocaleTimeString("en-GB", timeOptions);

        return `${formattedDate} ${formattedTime}`;
    }

    updateCurrentUserData = (userData) => {
        this.currentSessionUserData = userData;
    }
}
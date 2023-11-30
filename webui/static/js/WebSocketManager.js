export default class WebSocketManager {
    constructor() {
        this.socket = new WebSocket("ws://" + document.location.host + "/ws");
        this.initialize();
        this.eventListeners = {};
    }

    initialize() {
        this.socket.onopen = () => console.log("Websocket connection established");
        this.socket.onclose = () => console.log("Websocket connection closed");
        this.socket.onerror = (error) => console.log("WebSocket Error", error);
        this.socket.onmessage = this.handleMessages;
    }

    // ------------------------- INCOMING MESSAGES ----------------------------  

    //Function to register callback functions for incoming messages (by message type)
    on(event, callback) {
        if (!this.eventListeners[event]) {
            this.eventListeners[event] = [];
        }
        this.eventListeners[event].push(callback);
    }

    //If message is received, executes all functions which are binded to given message type
    emit(event, data) {
        if (this.eventListeners[event]) {
            this.eventListeners[event].forEach(callback => callback(data));
        }
    }

    //If receiving message(s), parse them and send to binded functions for execution
    handleMessages = (event) => {

        //Split by delimiter and remove any empty string in array with filter
        const rawMessages = event.data.split("\n").filter(Boolean);

        // Parse each message
        rawMessages.forEach(rawMessage => {
            try {
                const message = JSON.parse(rawMessage);
                this.emit(message.type, message.payload);
            } catch (error) {
                console.error("Unable to parse JSON string:", error);
            }
        });
    }

    // ------------------------- OUTGOING MESSAGES ----------------------------  

    sendLoginRequest(username, password) {
        this.socket.send(JSON.stringify({ Type: 'loginRequest', Payload: { "username": username, "password": password } }));
    }

    sendRegisterRequest(username, firstName, lastName, birthDate, gender, email, password) {
        this.socket.send(JSON.stringify({ 
            Type: 'registerRequest', 
            Payload: {
                "username": username,
                "firstName": firstName,
                "lastName": lastName,
                "dateBirth": birthDate,
                "gender": gender,
                "email": email,
                "password": password
            }}));
    }

    sendFullPostRequest(postId) {
        this.socket.send(JSON.stringify({ Type: 'fullPostAndCommentsRequest', Payload: postId }));
    }

    sendCreateNewPostRequest(date, postTheme, postContent, postCategories) {
        this.socket.send(JSON.stringify({ Type: 'newPostRequest', Payload: { "date": date, "theme": postTheme, "content": postContent, "categoriesID": postCategories } }));
    }

    sendLogOutRequest() {
        this.socket.send(JSON.stringify({ Type: 'logoutRequest' }));
    }

    sendGetPostsPortionRequest(afterID) {
        this.socket.send(JSON.stringify({ Type: 'postsPortionRequest', Payload: afterID }));
    }

    sendNewCommentSubmit(date, postID, commentContent) {
        this.socket.send(JSON.stringify({ Type: 'newCommentRequest', Payload: { "date": date, "post_id": postID, "content": commentContent } }));
    }

    sendOpenChatRequest(recipientUserID) {
        this.socket.send(JSON.stringify({ Type: 'openChatRequest', Payload: recipientUserID }));
    }

    sendCloseChatRequest() {
        this.socket.send(JSON.stringify({ Type: 'closeChatRequest' }));
    }

    sendChatMessage(date, messageContent) {
        this.socket.send(JSON.stringify({ Type: 'sendMessageToOpendChatRequest', Payload: { "date": date, "messageContent": messageContent } }));
    }

    sendChatPortionRequest(beforeID) {
        this.socket.send(JSON.stringify({ Type: 'chatPortionRequest', Payload: beforeID }));
    }
}
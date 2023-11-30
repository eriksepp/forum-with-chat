import { throttleAndDebounce } from "./helpers.js";
import { STRINGS } from "./ConstantStrings.js";

export default class PostsListView {
    constructor(DOMElements, switchChildView, webSocketManager) {
        this.DOMElements = DOMElements;
        this.switchChildView = switchChildView;
        this.webSocketManager = webSocketManager;

        this.throttleAndDebouncedHandleScroll = throttleAndDebounce(this.handleScroll.bind(this), 300);

        this.webSocketManager.on("postsPortionReply", this.handleReceivedPostsPortion);

        this.bottomPostID = 0;
        this.lastScrollPos = 0;
        this.allPostsLoaded = false;
        this.showMorePostsDelay = 0;
        this.gettingMorePosts = false;
    }

    show() {
        this.DOMElements.postsListContainer.style.display = "block";
        window.scrollTo({ left: 0, top: this.lastScrollPos, behavior: "instant" });
        this.bindEventListeners();
    }

    hide() {
        this.lastScrollPos = window.scrollY || document.documentElement.scrollTop;
        this.DOMElements.postsListContainer.style.display = "none";
        this.unbindEventListeners();
    }

    // ---------------------------- EVENT LISTENERS ---------------------------

    bindEventListeners = () => {
        this.DOMElements.postsListContainer.addEventListener("click", this.handleFullPostClick);
        this.DOMElements.postsListContainer.addEventListener("mouseover", this.handlePostHover);
        this.DOMElements.postsListContainer.addEventListener("mouseout", this.handlePostHoverOut);
        window.addEventListener("scroll", this.throttleAndDebouncedHandleScroll);
    }

    unbindEventListeners = () => {
        this.DOMElements.postsListContainer.removeEventListener("click", this.handleFullPostClick);
        this.DOMElements.postsListContainer.removeEventListener("mouseover", this.handlePostHover);
        this.DOMElements.postsListContainer.removeEventListener("mouseout", this.handlePostHoverOut);
        window.removeEventListener("scroll", this.throttleAndDebouncedHandleScroll);
    }

    // --------------------------- SCROLL HANDLING ----------------------------

    handleScroll = () => {
        //Handle pull-to-refresh (user scrolls up on top of page)
        let currentScrollPos = window.scrollY || document.documentElement.scrollTop;
        if (currentScrollPos <= 0) {
            this.emptyPostsListAndGetTenNewestPosts();
            return
        }

        //If it wasn't pull-to-refresh, check if all posts are loades
        if (this.allPostsLoaded === true) {
            return
        }

        //If all posts are not loaded, check if user is below enough to get more posts
        const offset = 600;
        if (!this.gettingMorePosts && window.innerHeight + window.scrollY >= document.body.offsetHeight - offset) {
            this.webSocketManager.sendGetPostsPortionRequest(this.bottomPostID);
            this.gettingMorePosts = true;
        }
    }

    // ------------------- RESET FEED, GET 10 LATEST POSTS --------------------

    resetPostsList = () => {
        this.lastScrollPos = 0;
        this.allPostsLoaded = false;
        this.showMorePostsDelay = 0;
        this.DOMElements.postsListContainer.innerHTML = "";
    }

    requestTenNewestPosts = () => {
        this.webSocketManager.sendGetPostsPortionRequest(0);
    }

    emptyPostsListAndGetTenNewestPosts() {
        this.showMorePostsDelay = 0;
        this.resetPostsList();
        this.requestTenNewestPosts();
    }

    // -------------- RECEIVING POSTS FROM SERVER & SHOWING THEM --------------

    handleReceivedPostsPortion = (payload) => {
        setTimeout(() => {
            if (payload.result !== STRINGS.SUCCESS) {
                return
            }

            //Remove previous loading spinner if it exists
            this.removeLoadingSpinner();

            //In case of no posts received, user has reached the end of feed
            if (payload.data === null) {
                return
            }

            //Parse each post data to variables and add post to DOM
            for (let i = 0; i < payload.data.length; i++) {
                const parsedPostData = this.parsePostData(payload.data[i]);
                this.addPostToDOM(parsedPostData);

                //Store last posts ID for further requests to server
                if (i === payload.data.length - 1) {
                    this.bottomPostID = parsedPostData.postID;
                };
            };
            
            //Add delay for further messages to avoid browser scroll momentum
            //(otherwise it keeps scrolling by itself after new messages have loaded)
            this.showMorePostsDelay = 600;

            this.manageStatusForGettingMorePosts(payload.data);

            this.gettingMorePosts = false;
            
        }, this.showMorePostsDelay);
    }

    //If at least 10 new posts received, there may be more. If less than 10, user has arrived to end
    manageStatusForGettingMorePosts = (newPosts) => {
        if (newPosts.length >= 10) {
            this.showLoadingSpinner();
        } else {
            this.allPostsLoaded = true;
            this.showEndOfPostsMessage();
        }
    }

    //For adding to DOM, parse payload data to variables
    parsePostData = (postData) => {
        return {
            postID: postData.id,
            postLikesQty: postData.message.likes[0],
            postDislikesQty: postData.message.likes[1],
            postCommentsQty: postData.commentsQuantity ? postData.commentsQuantity : 0,
            postTheme: postData.theme,
            postAuthor: postData.message.author.name,
            postContent: postData.message.content,
        };
    }

    generatePostHTML = (postData) => {
        return `
        <div class="card mb-3 border-0 pb-3">
            <div class="card-header d-flex bg-transparent border-0 ps-0 pb-0 justify-content-between">
                <div class="left-section d-flex align-items-center">
                    <span class="post-title fw-bolder fs-5 text-decoration-none me-2" data-id="${postData.postID}">
                        ${postData.postTheme}</span>
                </div>
                <div class="right-section d-flex flex-nowrap align-items-center">
                    <span class="username me-3">${postData.postAuthor}</span>
                    <i class="fa-regular fa-comment fa-xl me-1"></i>
                    <span class="comment-count">${postData.postCommentsQty}</span>
                </div>
            </div>
            <div class="post-text-wrapper card-body text-decoration-none ps-0 pt-1" data-id="${postData.postID}">
                <p class="post-text mb-0">
                    ${postData.postContent}
                </p>
            </div>
        </div>`
    }

    addPostToDOM = (postData) => {
        const postHTML = this.generatePostHTML(postData);
        this.DOMElements.postsListContainer.innerHTML += postHTML;
    }

    // ------------- LOADING SPINNER AND END OF POST FEED MESSAGE -------------

    //Show spinner below last post while loading more posts from server
    showLoadingSpinner = () => {
        const spinnerHTML = `
        <div id="postsListLoadingSpinner" class="d-flex justify-content-center mb-5">
            <div class="spinner-border text-primary" role="status"></div>
        </div>`
        this.DOMElements.postsListContainer.innerHTML += spinnerHTML;
    }

    //Hide loading spinner when new posts arrived or no more posts to show
    removeLoadingSpinner = () => {
        const spinnerEl = document.getElementById("postsListLoadingSpinner");
        if (spinnerEl) {
            spinnerEl.remove();
        }
    }

    //Show a message in the bottom of posts feed when user has seen all existing posts
    showEndOfPostsMessage = () => {
        const messageHTML = `<span class="d-block text-center mb-5">You've read all there is to read.</span>`;
        this.DOMElements.postsListContainer.innerHTML += messageHTML;
    }

    // ------------------------ POST TITLE/TEXT HOVER -------------------------
    //If user hovers on post title then also post text changes color and vice versa

    handlePostHover = (event) => {
        const target = event.target.closest(".post-title, .post-text-wrapper");
        if (target) {
            this.togglePostHoverEffect(target, true);
        }
    }

    handlePostHoverOut = (event) => {
        const target = event.target.closest(".post-title, .post-text-wrapper");
        if (target) {
            this.togglePostHoverEffect(target, false);
        }
    }

    togglePostHoverEffect(target, add) {
        let relatedElement;
        const postContainer = target.closest(".card");
        if (target.classList.contains("post-title")) {
            relatedElement = postContainer.querySelector(".post-text-wrapper");
        } else {
            relatedElement = postContainer.querySelector(".post-title");
        }

        if (relatedElement) {
            add ? relatedElement.classList.add("hovered") : relatedElement.classList.remove("hovered");
        }
    }

    // ------- USER CLICKS ON SOME POST TO SEE FULL POST WITH COMMENTS --------

    handleFullPostClick = (event) => {
        const target = event.target.closest(".post-title, .post-text-wrapper");
        if (target) {
            this.togglePostHoverEffect(target, false);
            const postId = target.dataset.id;
            this.webSocketManager.sendFullPostRequest(Number(postId));
        }
    }
}
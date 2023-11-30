import { getCurrentISODate } from "./helpers.js";
import { STRINGS } from "./ConstantStrings.js";

export default class FullPostView {
    constructor(DOMElements, switchChildView, webSocketManager) {
        this.DOMElements = DOMElements;
        this.switchChildView = switchChildView;
        this.webSocketManager = webSocketManager;
        this.webSocketManager.on("newCommentReply", this.handleNewCommentReply);
    }

    show() {
        this.DOMElements.fullPostContainer.style.display = 'block';
        this.bindEventListeners();
    }

    hide() {
        this.DOMElements.fullPostContainer.style.display = 'none';
        this.unbindEventListeners();
        this.resetCommentForm();
    }

    // ---------------------------- EVENT LISTENERS ---------------------------

    bindEventListeners = () => {
        this.DOMElements.newCommentInput.addEventListener("click", this.expandCommentForm);
        this.DOMElements.cancelNewComment.addEventListener("click", this.resetCommentForm);
        this.DOMElements.newCommentForm.addEventListener("submit", this.handleSubmittingNewComment);
    }

    unbindEventListeners = () => {
        this.DOMElements.newCommentInput.removeEventListener("click", this.expandCommentForm);
        this.DOMElements.cancelNewComment.removeEventListener("click", this.resetCommentForm);
        this.DOMElements.newCommentForm.removeEventListener("submit", this.handleSubmittingNewComment);
    }

    // -------------------- SHOWING FULL POST AND COMMENTS --------------------

    //Show full post and comments received from server
    setFullPostData(payload) {
        if (payload.result !== STRINGS.SUCCESS) {
            return
        }
        const { id, theme, message: { author: { name: username }, content: content }, commentsQuantity, categories, comments } = payload.data;

        this.updatePostMainData(id, theme, username, commentsQuantity, content);
        this.updatePostCategories(categories);
        this.updatePostComments(comments);
    }

    //Update id, title, username, comments amount and post text
    updatePostMainData = (id, theme, username, commentsQuantity, content) => {
        this.DOMElements.fullPostIdForComment.value = id;
        this.DOMElements.fullPostTitle.textContent = theme;
        this.DOMElements.fullPostUsername.textContent = username;
        this.DOMElements.fullPostCommentAmount.textContent = commentsQuantity;
        this.DOMElements.fullPostContent.textContent = content;
    }

    //Remove old categories and append new post ones
    updatePostCategories = (categories) => {
        this.DOMElements.fullPostCategoriesWrapper.innerHTML = "";
        categories.forEach(({ id, name }) => {
            const categoryDiv = document.createElement("div");
            categoryDiv.classList.add("categoryBubble");
            categoryDiv.dataset.id = id;
            categoryDiv.textContent = name;
            this.DOMElements.fullPostCategoriesWrapper.appendChild(categoryDiv);
        });
    }

    //Remove old comments and add new ones if any
    updatePostComments = (comments) => {
        this.DOMElements.fullPostCommentsContainer.innerHTML = "";
        if (comments) {
            comments.reverse().forEach(({ message: { author: { name: commentAuthor }, content: commentText } }) => {
                this.fullPostAddCommentToDOM(commentAuthor, commentText);
            })
        }
    }

    //Creates DOM elements with author and text for a comment
    fullPostAddCommentToDOM = (commentAuthor, commentText) => {
        const commentHTML = this.generateCommentHTML(commentAuthor, commentText)
        this.DOMElements.fullPostCommentsContainer.insertAdjacentHTML("beforeend", commentHTML);
    }

    generateCommentHTML = (commentAuthor, commentText) => {
        return `<div class="comment-box card mb-3 p-3 pt-1 border-0">
                    <div class="card-header d-flex bg-transparent border-0 ps-0 pb-1 pe-0 justify-content-between">
                        <div class="left-section d-flex align-items-center">
                            <span class="fw-semibold fs-7">${commentAuthor}:</span>
                        </div>
                    </div>
                    <p class="mb-0">${commentText}</p>
                </div>`
    }

    // --------------------------- NAVIGATING AWAY ----------------------------

    handleBackArrowClick = () => {
        this.switchChildView("postsList");
    }

    // --------------------- COMMENT FORM EXPAND/COLLAPSE ---------------------

    //To expand and collapse comment form when users focuses/unfocuses it

    expandCommentForm = () => {
        this.DOMElements.newCommentInput.style.height = "20vh";
        this.DOMElements.commentFormButtons.classList.remove("d-none");
        this.scrollElementIntoViewIfNeeded(commentFormButtons);
    }

    collapseCommentForm = () => {
        this.DOMElements.newCommentInput.style.height = "calc(3.5rem + 2px)";
        this.DOMElements.commentFormButtons.classList.add("d-none");
    }

    // ------------------------ SUBMITTING NEW COMMENT ------------------------

    //Send new comment to server on submit
    handleSubmittingNewComment = (event) => {
        event.preventDefault();
        const postID = Number(this.DOMElements.fullPostIdForComment.value);
        const commentText = this.DOMElements.newCommentInput.value;
        if (commentText.trim() === "") {
            this.DOMElements.newCommentInput.classList.add("is-invalid")
            document.getElementById("newCommentTextLabel").innerHTML = "Add some text to your comment";
            return
        }
        const date = getCurrentISODate();
        this.webSocketManager.sendNewCommentSubmit(date, postID, commentText);
    }

    // ------------------------- SHOWING NEW COMMENT --------------------------

    //After submitting new comment, server sends whole post and all comments again
    handleNewCommentReply = (payload) => {
        if (payload.result !== STRINGS.SUCCESS) {
            console.log("Error: Creating new comment failed")
            return
        }
        this.setFullPostData(payload);
        this.resetCommentForm();
        const newCommentEl = this.DOMElements.fullPostCommentsContainer.firstElementChild;
        this.scrollElementIntoViewIfNeeded(newCommentEl);
    }

    resetCommentForm = () => {
        this.DOMElements.newCommentInput.value = "";
        this.collapseCommentForm();
    }

    //After submitting new comment, scroll it into view so user can see it
    scrollElementIntoViewIfNeeded = (element) => {
        const elementBottomPos = element.getBoundingClientRect().bottom;
        
        if (elementBottomPos > window.innerHeight) {
            element.scrollIntoView({ block: "end", behavior: "smooth" });
        }
    }
}
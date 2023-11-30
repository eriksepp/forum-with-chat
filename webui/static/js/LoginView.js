export default class LoginView {
  constructor(DOMElements, switchView, webSocketManager) {
    this.DOMElements = DOMElements;
    this.switchView = switchView;
    this.webSocketManager = webSocketManager;
  }

  show() {
    this.DOMElements.loginRegisterWrapper.style.display = "block";
    this.DOMElements.loginContainer.classList.add("d-flex");
    this.bindEventListeners();
  }

  hide() {
    this.DOMElements.loginRegisterWrapper.style.display = "none";
    this.DOMElements.loginContainer.classList.remove("d-flex");
    this.DOMElements.loginContainer.style.display = "none";
    document.getElementById("logoutSuccessMessage").setAttribute('hidden', '');
    this.unbindEventListeners();
    this.unfocusLoginFields();
    this.clearLoginForm();
  }

  // ----------------------------- EVENT LISTENERS ----------------------------

  bindEventListeners = () => {
    this.DOMElements.loginForm.addEventListener("submit", this.handleLoginFormSubmit);
    this.DOMElements.goToRegisterButton.addEventListener("click", this.handleGoToRegisterClick);
  }

  unbindEventListeners = () => {
    this.DOMElements.loginForm.removeEventListener("submit", this.handleLoginFormSubmit);
    this.DOMElements.goToRegisterButton.removeEventListener("click", this.handleGoToRegisterClick);
  }

  // ----------------------------- LOADING SCREEN -----------------------------

  //Loading screen is shown before login form until server tells if active session exists or not
  hideLoadingScreen = () => {
    this.DOMElements.loginTitle.style.transition = "none";
    this.DOMElements.loginWelcomeToText.style.transition = "none";
    this.DOMElements.loginForm.style.transition = "none";
    this.DOMElements.goToRegisterButton.style.transition = "none";
    this.DOMElements.loginSpinner.style.display = "none";
    this.DOMElements.loginTitle.style.marginTop = "-12px";
    this.DOMElements.loginWelcomeToText.style.opacity = "1";
    this.DOMElements.loginForm.style.opacity = "1";
    this.DOMElements.goToRegisterButton.style.opacity = "1";
  }

  // ------------------------------- LOGIN FORM -------------------------------

  showLoginForm = () => {
    this.DOMElements.loginSpinner.style.display = "none";
    this.DOMElements.loginTitle.style.marginTop = "-12px";
    //Timeouts for animation
    setTimeout(() => {
      this.DOMElements.loginForm.style.visibility = "visible";
      this.DOMElements.loginWelcomeToText.style.opacity = "1";
      this.DOMElements.loginForm.style.opacity = "1";
      this.DOMElements.goToRegisterButton.style.opacity = "1";

      //For delaying browser autologin pop-up showing before animation is done
      setTimeout(() => {
        document.getElementById("loginUsername").focus();
      }, 200);
      
    }, 400)
  }

  //If user logs out, different message than on first visit is shown
  showLoggedOutView = () => {
    this.hideLoadingScreen();
    this.DOMElements.loginForm.style.visibility = "visible";
    document.getElementById("loginWelcomeToTxt").setAttribute('hidden', '');
    document.getElementById("logoutSuccessMessage").removeAttribute("hidden");
  }

  handleLoginFormSubmit = (event) => {
    event.preventDefault();
    const usernameField = document.getElementById("loginUsername");
    const passwordField = document.getElementById('loginPassword');

    usernameField.classList.remove("is-invalid");
    passwordField.classList.remove("is-invalid");

    const username = usernameField.value;
    const password = passwordField.value;

    //Check if both fields are filled
    if (username.trim() === "") {
      this.showLoginFormError("loginUsername", "Username missing");
      return
    } else if (password.trim() === "") {
      this.showLoginFormError("loginPassword", "Password missing");
      return
    }
    
    this.webSocketManager.sendLoginRequest(username, password);
  }

  //To handle server's error message after login attempt
  handleLoginError = (payload) => {
    if (payload.data.startsWith("user")) {
      this.showLoginFormError("loginUsername", payload.data);
    } else {
      this.showLoginFormError("loginPassword", payload.data);
    }
  }

  //Show login error message on according field
  showLoginFormError = (fieldID, message) => {
    document.getElementById("logoutSuccessMessage").setAttribute('hidden', '');
    const loginErrorMessageField = document.getElementById("loginErrorMessage");
    loginErrorMessageField.textContent = message;
    loginErrorMessageField.removeAttribute('hidden');
    const passwordField = document.getElementById('loginPassword');
    passwordField.value = "";
    if (fieldID === "loginUsername") {
      const usernameField = document.getElementById('loginUsername');
      usernameField.classList.add("is-invalid");
      usernameField.value = "";
      usernameField.focus();
    } else {
      passwordField.classList.add("is-invalid");
    }
  }

  // ---------------------------- NAVIGATING AWAY -----------------------------

  handleGoToRegisterClick = () => {
    this.switchView("register");
  }

  //If navigating away from login form, unfocus login fields
  unfocusLoginFields = () => {
    document.getElementById("loginUsername").blur();
    document.getElementById('loginPassword').blur();
  }

  //If navigating away, remove empty login form and remove error messages
  clearLoginForm() {
    document.getElementById('loginForm').reset();
    const loginErrorMessageField = document.getElementById('loginErrorMessage');
    loginErrorMessageField.setAttribute('hidden', '');
    document.getElementById('loginUsername').classList.remove("is-invalid");
    document.getElementById('loginPassword').classList.remove("is-invalid");
  }
}
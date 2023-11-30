import { DOMElements } from "./DOMElements.js";
import { STRINGS } from "./ConstantStrings.js";
import WebSocketManager from "./WebSocketManager.js";
import RegisterView from "./RegisterView.js";
import LoginView from "./LoginView.js";
import DashboardView from "./DashboardView.js";

export default class RootViewManager {
  constructor() {
    this.DOMElements = DOMElements;
    this.webSocketManager = new WebSocketManager();

    this.webSocketManager.on("currentSession", this.handleExistingSessionMessage);
    this.webSocketManager.on("loginReply", this.handleLoginReply);
    this.webSocketManager.on("registerReply", this.handleRegisterReply)
    this.webSocketManager.on("logoutReply", this.handleLogOut);

    this.views = {
      login: new LoginView(this.DOMElements, this.switchView, this.webSocketManager),
      register: new RegisterView(this.DOMElements, this.switchView, this.webSocketManager),
      dashboard: new DashboardView(this.DOMElements, this.switchView, this.webSocketManager),
    };

    this.currentView = null;

    this.init();
  }

  //After creating class, show login view
  init = () => {
    this.switchView(STRINGS.LOGIN);
  }

  //Main function to switch between login, register, dashboard view
  switchView = (newView) => {
    if (this.currentView) {
      this.currentView.hide();
    }
    this.currentView = this.views[newView];
    this.currentView.show();
  }

  // -------------------- HANDLE CURRENT SESSION STATUS -----------------------

  //After ws connection established, server will send info if user session exists
  handleExistingSessionMessage = (payload) => {
    if (payload.result !== STRINGS.SUCCESS) {
      return
    }
    
    //Timeout to not shown loading screen for just a microsecond to user if server sends msg very quickly
    setTimeout(() => {
      if (payload.data.user !== null) {
        this.handleLoginReply(payload);
      } else {
        this.views.login.showLoginForm();
      }
    }, 250)
  };

  // ------------------------------ HANDLE LOGIN ------------------------------

  handleLoginReply = (payload) => {
    if (payload.result === STRINGS.SUCCESS) {
      this.handleSuccessfulLogin(payload);
    } else {
      this.views.login.handleLoginError(payload);
    }
  };

  handleSuccessfulLogin = (payload) => {
    this.storeSessionAsCookie(payload.data);
    this.views.dashboard.storeCurrentSessionUserData(payload.data.user);
    this.views.dashboard.showUsernameInNavbar();
    this.switchView(STRINGS.DASHBOARD);
  };

  storeSessionAsCookie = (data) => {
    const uuid = data.user.uuid;
    const expiryTime = (new Date(data.user.expirySession)).toUTCString();
    document.cookie = `forum_session_id=${uuid}; expires=${expiryTime}`;
  }

  // ----------------------------- HANDLE LOGOUT ------------------------------

  //After successful logout reply, remove cookie, reset views and show login screen
  handleLogOut = (payload) => {
    if (payload.result === STRINGS.SUCCESS) {
      this.views.dashboard.removeChatRecipientData();
      document.cookie = "forum_session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
      this.views.login.showLoggedOutView();
      this.switchView(STRINGS.LOGIN);
    }
  }

  // --------------------------- HANDLE REGISTER ------------------------------

  //In case of successful register, log user in. If unsuccessful, show error on register form
  handleRegisterReply = (payload) => {
    if (payload.result === STRINGS.SUCCESS) {
      this.handleSuccessfulLogin(payload);
    } else {
      this.views.register.handleRegisterError(payload);
    }
  };
}
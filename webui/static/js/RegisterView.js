export default class RegisterView {
  constructor(DOMElements, switchView, webSocketManager) {
      this.DOMElements = DOMElements;
      this.switchView = switchView;
      this.webSocketManager = webSocketManager;
      this.createBirthdayOptions();
  }

  show() {
    this.DOMElements.loginRegisterWrapper.style.display = "block";
    this.DOMElements.registerContainer.classList.add("d-flex");
    this.DOMElements.registerContainer.style.display = "block";
    document.getElementById("registerUsername").focus();
    this.bindEventListeners();
    this.enablePeekPassword();
  }

  hide() {
    this.DOMElements.loginRegisterWrapper.style.display = "none";
    this.DOMElements.registerContainer.classList.remove("d-flex");
    this.DOMElements.registerContainer.style.display = "none";
    this.unbindEventListeners();
    this.disablePeekPassword();
    this.clearRegisterForm();
  }

  // ----------------------------- EVENT LISTENERS ----------------------------

  bindEventListeners = () => {
    this.DOMElements.registerForm.addEventListener("submit", this.handleRegisterFormSubmit);
    this.DOMElements.goToLoginButton.addEventListener("click", this.handleBackToLoginClick);
    this.DOMElements.registerTitle.addEventListener("click", this.handleBackToLoginClick);
    document.getElementById("registerEmail").addEventListener("invalid", this.preventBrowserDefaultInvalidEmailResponse);
  }

  unbindEventListeners = () => {
    this.DOMElements.registerForm.removeEventListener("submit", this.handleRegisterFormSubmit);
    this.DOMElements.goToLoginButton.removeEventListener("click", this.handleBackToLoginClick);
    this.DOMElements.registerTitle.removeEventListener("click", this.handleBackToLoginClick);
    document.getElementById("registerEmail").removeEventListener("invalid", this.preventBrowserDefaultInvalidEmailResponse);
  }

  // ------------------------ INITALIZE REGISTER VIEW -------------------------

  //Create option elements for choosing a birthday
  createBirthdayOptions() {
    const birthDayContainer = document.getElementById("registerBirthDay");
    for (let i = 1; i<=31; i++) {
      const dateOption = document.createElement("option");
      dateOption.value = i;
      if (i < 10) {
        dateOption.value = "0" + dateOption.value;
      }
      dateOption.textContent = i;
      birthDayContainer.appendChild(dateOption);
    }
  
    const birthMonthContainer = document.getElementById("registerBirthMonth");
    const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
    for (let i = 1; i<=12; i++) {
      const monthOption = document.createElement("option");
      monthOption.value = i;
      if (i < 10) {
        monthOption.value = "0" + monthOption.value;
      }
      monthOption.textContent = months[i-1];
      birthMonthContainer.appendChild(monthOption);
    }
  
    const birthYearContainer = document.getElementById("registerBirthYear");
    for (let i = 2009; i >= 1923; i--) {
      const yearOption = document.createElement("option");
      yearOption.value = i;
      yearOption.textContent = i;
      birthYearContainer.appendChild(yearOption);
    }
  }

  // ----------------------------- EVENT HANDLERS -----------------------------

  handleBackToLoginClick = () => {
    this.switchView("login");
  }

  handleRegisterFormSubmit = (event) => {
    event.preventDefault();
    this.resetPreviousInvalidFormFields();

    const registerFormData = {
      username: document.getElementById("registerUsername").value,
      firstName: document.getElementById("registerFirstName").value,
      lastName: document.getElementById("registerLastName").value,
      birthDateDay: document.getElementById("registerBirthDay").value,
      birthDateMonth: document.getElementById("registerBirthMonth").value,
      birthDateYear: document.getElementById("registerBirthYear").value,
      gender: document.getElementById("registerGender").value,
      email: document.getElementById("registerEmail").value,
      password: this.DOMElements.passwordFieldForRegister.value,
    }

    const error = this.validateRegisterFormData(registerFormData);

    if (error) {
      this.showError(error.field, error.message);
      return;
    }

    const birthDate = registerFormData.birthDateYear + "-" + registerFormData.birthDateMonth + "-" + registerFormData.birthDateDay;

    this.webSocketManager.sendRegisterRequest(registerFormData.username, 
                                              registerFormData.firstName, 
                                              registerFormData.lastName, 
                                              birthDate, 
                                              registerFormData.gender, 
                                              registerFormData.email, 
                                              registerFormData.password);
  }

  handleRegisterError = (payload) => {
    document.getElementById("registerTitle").style.marginBottom = "-1vh";
    const registerErrorMessageField = document.getElementById("registerErrorMessage");
    registerErrorMessageField.removeAttribute('hidden');
    registerErrorMessageField.textContent = payload.data;
  }

  enablePeekPassword = () => {
    // Attach listeners as properties of the DOM element
    this.DOMElements.peekPasswordIcon.showPassword = () => this.togglePasswordFieldPeeking('show');
    this.DOMElements.peekPasswordIcon.hidePassword = () => this.togglePasswordFieldPeeking('hide');

    this.DOMElements.peekPasswordIcon.addEventListener('mouseover', this.DOMElements.peekPasswordIcon.showPassword);
    this.DOMElements.peekPasswordIcon.addEventListener('mouseout', this.DOMElements.peekPasswordIcon.hidePassword);
  }

  disablePeekPassword = () => {
    this.DOMElements.peekPasswordIcon.removeEventListener('mouseover', this.DOMElements.peekPasswordIcon.showPassword);
    this.DOMElements.peekPasswordIcon.removeEventListener('mouseout', this.DOMElements.peekPasswordIcon.hidePassword);
  }

  //Helper function to enable and disable peeking password field when registering
  togglePasswordFieldPeeking = (action) => {
    if (action === 'show') {
      this.DOMElements.passwordFieldForRegister.type = 'text';
    } else if (action === 'hide') {
      this.DOMElements.passwordFieldForRegister.type = 'password';
    }
  }

  preventBrowserDefaultInvalidEmailResponse = (e) => {
    e.preventDefault();
    this.showError("registerEmail", "E-mail address is not valid");
  }

  // --------------- REGISTER FORM VALIDATION & ERROR MESSAGES ----------------

  validateRegisterFormData = (formData) => {
    const emailPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    if (formData.username.trim() === "") {
      return { field: "registerUsername", message: "Username missing" };
    } else if (formData.firstName.trim() === "") {
      return { field: "registerFirstName", message: "First name missing" };
    } else if (formData.lastName.trim() === "") {
      return { field: "registerLastName", message: "Last name missing" };
    } else if (!formData.birthDateDay) {
      return { field: "registerBirthDay", message: "Birthday day missing" };
    } else if (!formData.birthDateMonth) {
      return { field: "registerBirthMonth", message: "Birthday month missing" };
    } else if (!formData.birthDateYear) {
      return { field: "registerBirthYear", message: "Birthday year missing" };
    } else if (!this.isValidDate(formData.birthDateDay, formData.birthDateMonth, formData.birthDateYear)) {
      return { field: "allBirthFields", message: "Birthday is not a valid date" };
    } else if (!formData.gender) {
      return { field: "registerGender", message: "No gender selected" };
    } else if (!emailPattern.test(formData.email)) {
      return { field: "registerEmail", message: "E-mail address is not valid" };
    } else if (formData.password.trim() === "") {
      return { field: "registerPassword", message: "Password missing" };
    } else if (formData.password.trim().length < 8) {
      return { field: "registerPassword", message: "Password must be at least 8 characters" };
    }
    return null;
  }

  showError = (fieldID, message) => {
    document.getElementById("registerTitle").style.marginBottom = "-1vh";
    const registerErrorMessageField = document.getElementById("registerErrorMessage");
    registerErrorMessageField.removeAttribute('hidden');
    registerErrorMessageField.textContent = message;

    //In case of invalid date, mark all birthday fields red
    if (fieldID === "allBirthFields") {
      document.getElementById("registerBirthDay").style.borderColor = "#dc3545";
      document.getElementById("registerBirthMonth").style.borderColor = "#dc3545";
      document.getElementById("registerBirthYear").style.borderColor = "#dc3545";
      return;
    }

    const inputField = document.getElementById(fieldID);
    if (fieldID.startsWith("registerBirth")) {
      inputField.style.borderColor = "#dc3545"; //"Is-invalid" BS5 class does not look good on date dropdowns
    } else {
      inputField.classList.add("is-invalid");
    }
    inputField.value = "";
  }

  isValidDate(day, month, year) {
    const dateObject = new Date(year, month - 1, day);

    //Check if dateObject is valid and there is no date overflow
    return (
      dateObject &&
      dateObject.getFullYear() === parseInt(year) &&
      dateObject.getMonth() === parseInt(month-1) &&
      dateObject.getDate() === parseInt(day)
    )
  }

  // ------------------------ RESET REGISTER FORM -----------------------------    

  resetPreviousInvalidFormFields() {
    const formFields = document.querySelectorAll('.form-control, .form-select');
    formFields.forEach((element) => {
      element.style.borderColor = "";
      element.classList.remove('is-invalid')});
  }

  clearRegisterForm() {
    document.getElementById('registerForm').reset();
    document.getElementById("registerErrorMessage").setAttribute('hidden', '');
    document.getElementById("registerTitle").style.marginBottom = "0.5vh";
    this.resetPreviousInvalidFormFields();
  }
}
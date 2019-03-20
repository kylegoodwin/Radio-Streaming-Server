import { observable, decorate, action } from "mobx";

class GlobalStore {

    password = "";
    email = "";

    isAuthenticating = false;
    isAuthenticated =  false;

    isPlaying = true;

    context = new AudioContext();
    gainNode = this.context.createGain();
    
    socket;

    userHasAuthenticated = authenticated => {
        this.isAuthenticated = authenticated;
    }

    handleLogout = async event => {
        this.userHasAuthenticated(false);
      }

    validateForm = () => {
        return this.email.length > 0 && this.password.length > 0;
    }

    setEmail = email => {
        this.email = email;
    }

    setPassword = password => {
        this.password = password;
    }

    storeUserInfo = userObject => {
        this.username = userObject.username;
    }
};

decorate(GlobalStore, {
    socket: observable,
    isPlaying: observable,
    context: observable,
    username: observable,
    password: observable,
    userObject: observable,
    email: observable,
    token: observable,
    gainNode: observable,
    validateForm: action,
    setEmail: action,
    setPassword: action,
    isAuthenticating: observable,
    isAuthenticated: observable,
    handleLogout: action,
    userHasAuthenticated: action,
    storeUserInfo: action
});

export default new GlobalStore();
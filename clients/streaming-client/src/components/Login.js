import React, { Component } from "react";
import { FormGroup, FormControl } from "react-bootstrap";
import  { Redirect, Link } from 'react-router-dom';
import { observer } from "mobx-react"

// Components
import "./SignUp.css";
import LoaderButton from "../components/LoaderButton";

// Stores
import GlobalStore from "../GlobalStore"


class Login extends Component {
  constructor(props) {
    super(props);
    this.state = {alert: null}
  }

  componentDidMount = () => {
    document.body.style.backgroundColor = "rgb(32, 38, 44)";
  }

  componentWillUnmount = () => {
    document.body.style.backgroundColor = null;
  }

  handleEmailChange = event => {
    GlobalStore.setEmail(event.target.value);
  }

  handlePasswordChange = event => {
    GlobalStore.setPassword(event.target.value);
  }

  /*
  getUserInfo(){
      return API.get("posts", "/users");
  }

  setUserInfo(username, email){
    return API.post("posts", "/users", {
      body: {
        email: email,
        username: username
      }
    });
  }
  */

  loginUser(email, password){
    let credentials = {
        email: email,
        password: password
    }
    fetch('https://audio-api.kjgoodwin.me/v1/sessions',{
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                  "Accept": "application/json",
                },
                body: JSON.stringify(credentials)
            }).then((response) => {
              GlobalStore.token = response.headers.get('Authorization')
              GlobalStore.socket = 'wss://audio-api.kjgoodwin.me/ws?auth=' + GlobalStore.token
              return response.json()
            })
            .then((data)=>{
                GlobalStore.isAuthenticated = true;
              
                GlobalStore.currentUser = data;
                this.props.history.push("/home");
            });
  }

  handleSubmit = async event => {
    event.preventDefault();
    GlobalStore.isLoading = true;
  
    try {
      await this.loginUser(GlobalStore.email, GlobalStore.password);
      //GlobalStore.setAuthToken()
      console.log("done")
      this.props.history.push("/home");
      GlobalStore.email = "";
      GlobalStore.password = "";
    } catch (e) {
      this.setState({alert: e.message});
    }
    GlobalStore.isLoading = false;
  }
  

  render() {
    return (
      GlobalStore.isAuthenticated ?
      <Redirect to={{ pathname: '/home', state: { from: this.props.location } }} />
      :
      <div className="Login">
        {this.props.location.state && this.props.location.state.from && this.props.location.state.from.pathname === "/SignUp" ? <p className="signup-redirect" style={{color: "#FFF", padding: "1rem", fontSize: "16px"}}>Welcome to Produce! Check your email to verify your account.</p> : null}
        {this.state.alert ? <div className="signup-error">{this.state.alert}</div> : <div className="signup-no-error">t</div>}
        <form onSubmit={this.handleSubmit}>
          <FormGroup controlId="email" bsSize="large">
            <FormControl
              autoFocus
              type="text"
              value={GlobalStore.email}
              onChange={this.handleEmailChange}
              placeholder="Username or Email"
            />
          </FormGroup>
          <FormGroup controlId="password" bsSize="large">
            <FormControl
              value={GlobalStore.password}
              onChange={this.handlePasswordChange}
              type="password"
              placeholder="Password"
            />
          </FormGroup>
          <LoaderButton
            block
            bsSize="large"
            disabled={!GlobalStore.validateForm()}
            type="submit"
            isLoading={GlobalStore.isLoading}
            text="Login"
            className="login-button"
            loadingText="Logging in…"
          />
        </form>
        <Link className="switch-portals" to="/Signup">Signup</Link>
      </div>      
    );
  }
}

export default observer(Login);

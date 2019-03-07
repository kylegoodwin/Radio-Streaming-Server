import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import Home from './components/Home'
import {Link, Route} from 'react-router-dom';
import StartStream from './components/StartStream';

class App extends Component {
  constructor(props){
      super(props);
      this.state = {
          loggedIn: false,
          currentUser:{}
      };
  };

  loginUser = (credentials) =>{
    fetch('https://api.radio-stream.com/v1/sessions/', {
            method: "POST",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(credentials)
        }).then((response) => response.json())
        .then((data)=>{
            this.setState({currentUser:data, loggedIn: true});
            //redirect to home
            //this.props.history.push('/')
        });
  }

  render() {
    return (
      <Switch>
        <Route exact path='/' render={(routerProps) => (
                        <Home {...routerProps} loggedIn={this.state.loggedIn} currentUser={this.state.currentUser} />
                    )} />
        <Route path="/login" render={(routerProps) => (
                        <Login {...routerProps} loginUser={this.loginUser.bind(this)} loggedIn={this.state.loggedIn} />
                    )} />
        <Route path="/start-stream" component={StartStream} />
      </Switch>
    );
  }
}

export default App;

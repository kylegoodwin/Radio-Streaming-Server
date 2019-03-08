import React, {Component} from 'react';
import {Link, Route, Redirect} from 'react-router-dom';

class Login extends Component{
    constructor(props) {
        super(props);
        this.state = {
          email: "",
          password: "",
          currentUser: "",
          loggedIn:Boolean,
        };
    
        this.handleInputChange = this.handleInputChange.bind(this);
      }
    
      handleInputChange(event) {
        const target = event.target;
        const value = target.value
        const name = target.name;
    
        this.setState({
          [name]: value
        });
      }

      handleSubmit = () =>{
          //login user
        var credentials = {
            email: this.state.email,
            password: this.state.password
        }

        this.props.loginUser(credentials)
      }

      componentDidMount(){
          this.setState({loggedIn:this.props.loggedIn})
      }
    
      render() {
        if(loggedIn){
            return <Redirect to="/"/>
        } else {
            return (
            <form onSubmit={this.handleSubmit}>
                <label>
                Email
                <input
                    name="email"
                    type="text"
                    onChange={this.handleInputChange} />
                </label>
                <br />
                <label>
                Password
                <input
                    name="password"
                    type="text"
                    onChange={this.handleInputChange} />
                </label>
                <br />
                <input type="submit" value="Login" />
            </form>
            );
        }
      }
}

export default withRouter(Login)

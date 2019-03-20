import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import Home from './components/Home'
import {Link, Route, Switch} from 'react-router-dom';
import AppliedRoute from './components/AppliedRoute'
import StartStream from './components/StartStream';
import Login from './components/Login';
import AudioStream from './components/AudioStream';
import GlobalStore from './GlobalStore';

export default ({ childProps }) =>
<Switch>
    <AppliedRoute path="/" exact component={Home} props={childProps} />
    <AppliedRoute path="/login" exact component={Login} props={childProps} />
    <AppliedRoute path="/home" component={Home} props={childProps} />
    {/*<AppliedRoute path="/signup" exact component={Signup} props={childProps} />*/}
    <Route path="/start-stream" component={StartStream} />
    <Route path="/channels/" component={AudioStream} />
</Switch>;
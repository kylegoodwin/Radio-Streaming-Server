import React, {Component} from 'react';
import {Link, Route, Redirect} from 'react-router-dom';
import placeholder from '../img/stream-placeholder.png'

class StartStream extends Component{

    constructor(props){
        super(props);
        this.state = {
            currentUser: 0,
            streamName: "Stream",
        };
    };

    getUserInfo = () => {
        fetch('https://api.radio-stream.com/v1/me/',{
            method: "GET",
            headers: {
                'Content-Type': 'application/json'
            }
        }).then((response) => response.json())
        .then((data)=>{
            this.setState({currentUser:data})
        });
    };

    startStream = (name) =>{
        //start the stream
    }

    handleChange = (event) =>{
        this.setState({streamName:event.target.value})
    }

    handleSubmit = (event) =>{
        event.preventDefault();
        this.startStream(this.state.streamName)
    }

    componentDidMount(){
        this.getUserInfo();
    }

    render(){
        return(
            <form onSubmit={this.handleSubmit}>
                <label>
                    Stream Name:
                    <input type="text" value={this.state.streamName} onChange={this.handleChange} />
                </label>
                <input type="submit" value="Start Stream" />
            </form>
        )
    }
}


export default StartStream
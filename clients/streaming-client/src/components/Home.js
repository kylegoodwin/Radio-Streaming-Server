import React, {Component} from 'react';
import {Link, Route, Redirect} from 'react-router-dom';
import placeholder from '../img/stream-placeholder.png'

class Home extends Component{
    constructor(props){
        super(props);
        this.state = {
            streams:[],
            loggedIn:false
        };
    };

    componentDidMount(){
        //check session token
        this.setState({loggedIn:this.props.loggedIn})
        getStreams();
    };

    getStreams = () =>{
        fetch('https://api.radio-stream.com/v1/channels/',{
            method: "GET",
            headers: {
                'Content-Type': 'application/json'
            }
        }).then((response) => response.json())
        .then((data)=>{
            let retrievedStreams = data.map((streamID) => {
                return streamID
            });
            this.setState({streams:retrievedStreams});
        });
    };

    render(){
        if(!loggedIn){
            return <Redirect to="/login"/>
        }
        let streamsArray = this.state.streams.map((stream) => {
            return(
                <StreamListing streamID={stream} key={stream}/>
            );
        })
        return (
            <div className="stream-list">
                {streamsArray}
            </div>
        );
    };
};

class StreamListing extends component{
    constructor(props){
        super(props);
        this.state = {
            name: "",
            photoUrl: {placeholder},
            creatorName: "",
            creatorID: 0
        };
    };

    getStreamInfo = () =>{
        //placeholder api call
        fetch('https://api.radio-stream.com/v1/channels/' + this.props.channelID,{
            method: "GET",
            headers: {
                'Content-Type': 'application/json'
            }
        }).then((response) => response.json())
        .then((data)=>{
            this.setState({name:data.name, 
                            creatorName:data.creatorName, 
                            creatorID: data.creatorID});
        });
    };

    componentDidMount(){
        getStreamInfo();
    };

    render(){
        let path = "/channels/" + this.props.channelID;
        return(
            <Link className="stream-listing" to={path}>
                <img src={this.state.photoUrl} />
                <h1>{this.state.name}</h1>
            </Link>
        )
    }
}

export default Home;
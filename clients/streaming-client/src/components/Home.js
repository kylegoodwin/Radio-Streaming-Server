import React, {Component} from 'react';
import {Link, Route, Redirect, Switch} from 'react-router-dom';
import placeholder from '../img/stream-placeholder.png'
import muaz from '../img/muaz.jpeg'

class Home extends Component{
    constructor(props){
        super(props);
        this.state = {
            streams:[],
            loggedIn:true
        };
    };

    componentDidMount(){
        //check session token
        //this.setState({loggedIn:this.props.loggedIn})
        console.log("fucku")

        this.getStreams();
    };

    getStreams = () =>{
        try{
            fetch('https://audio-api.kjgoodwin.me/v1/audio/streams/',{
                method: "GET",
                headers: {
                    'Content-Type': 'application/json'
                }
            }).then((response) => response.json())
            .then((data)=>{
                console.log(data)
                let retrievedStreams = data.map((streamID) => {
                    return streamID
                });
                this.setState({streams:retrievedStreams});
            });
        } catch(e){
            console.log(e)
        }
    };

    render(){
        if(!this.state.loggedIn){
            return <Redirect to="/login"/>
        }
        let streamsArray = this.state.streams.map((stream) => {
            return(
                <StreamListing channelID={stream.channelID} name={stream.displayName} key={stream}/>
            );
        })
        return (
            <div className="stream-list">
                {streamsArray}
            </div>
        );
    };
};

class StreamListing extends Component{
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
        this.getStreamInfo();
    };

    render(){
        let path = "/channels/" + this.props.channelID;
        return(
            <Link className="stream-listing" to={path}>
                <img className="stream-img" src={muaz} />
                <div className="stream-listing-info">
                    <h1>{this.props.name}</h1>
                    <h2>Streamer</h2>
                </div>
                
            </Link>
        )
    }
}

export default Home;
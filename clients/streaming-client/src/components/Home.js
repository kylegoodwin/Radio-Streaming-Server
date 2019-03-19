import React, {Component} from 'react';
import {Link, Route, Redirect, Switch} from 'react-router-dom';
import placeholder from '../img/stream-placeholder.png'
import muaz from '../img/muaz.jpeg'
import GlobalStore from '../GlobalStore'
import {observer} from 'mobx-react'
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
        //this.setState({loggedIn:this.props.loggedIn})
        console.log("home mounted")
        if(GlobalStore.isAuthenticated){
            this.getStreams()
        }
        
    };

    getStreams = () =>{
        try{
            console.log("try")
            fetch('https://audio-api.kjgoodwin.me/v1/channels?live=true',{
                method: "GET",
                headers: {
                    'Authorization': GlobalStore.token,
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
        if(GlobalStore.isAuthenticated){
            let streamsArray = this.state.streams.map((stream) => {
                return(
                    <StreamListing 
                    channelID={stream.channelID} 
                    name={stream.displayName} 
                    key={stream} 
                    creator={stream.creator.userName}
                    img={stream.creator.photoURL}
                    />
                );
            })
            return (
                <div className="home">
                    <div className="stream-list">
                        {streamsArray}
                    </div>
                </div>
            );
        } else {
            return <Redirect to={{ pathname: '/login', state: { from: this.props.location } }} />
        }
    };
}

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
        //this.getStreamInfo();
    };

    render(){
        let path = "/channels/" + this.props.channelID;
        return(
            <Link className="stream-listing" to={path}>
                <img className="stream-img" src={this.props.img} />
                <div className="stream-listing-info">
                    <h1>{this.props.name}</h1>
                    <h2>{this.props.creator}</h2>
                </div>
                
            </Link>
        )
    }
}

export default observer(Home);
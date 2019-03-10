import React, {Component} from 'react';
import {Link, Route, Redirect, Switch} from 'react-router-dom';
import placeholder from '../img/stream-placeholder.png'
import muaz from '../img/muaz.jpeg'
import testfile from '../mp3/test.mp3'
import ReactAudioPlayer from 'react-audio-player';

class AudioStream extends Component{
    constructor(props){
        super(props);
        this.state = {
            comments: ["fuck", "fucking shit", "oh fuck", "fuck me"]
        }
    }
    

    render(){
        var commentList = this.state.comments.map((comment) => {
            return <Comment username="Muaz Kahn" content={comment}></Comment>
        })
        return(
            <div className='stream-and-comments'>
                <div className='streamy'>
                    <img className='stream-img' src={muaz}/>
                    <div className='stream-player-info'>
                        <h1 className='stream-title'>Stream Name</h1>
                        <h2>Creator</h2>
                        <ReactAudioPlayer
                            className='player'
                            src={testfile}
                            controls
                        />
                    </div>
                </div>
                <div className='comments'>
                    {commentList}
                </div>
            </div>
        )
    }
}

class Comment extends Component{

    render(){
        return(
            <div className='comment'>
                <div className='comment-user'>
                    <div className='cropper'><img className="comment-img" src={muaz}></img></div>
                    <div className='comment-username'><p>{this.props.username}</p></div>
                </div>
                <p>{this.props.content}</p>
                
            </div>
        )
    }
}

export default AudioStream;
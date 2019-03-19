import React, {Component} from 'react';
import muaz from '../img/muaz.jpeg'
import RTCMultiConnection from 'rtcmulticonnection'
import GlobalStore from '../GlobalStore';
import {observer} from 'mobx-react'
import Websocket from 'react-websocket';
import {FormControl, FormGroup} from 'react-bootstrap';


class AudioStream extends Component{
    constructor(props){
        super(props);
        this.state = {
            streamID: "",
            content:"",
            comments: [{
                body: "new message",
                creator: {
                    userName: "MUAZ",
                    photoURL: "https://www.gravatar.com/avatar/4f64c9f81bb0d4ee969aaf7b4a5a6f40",
                }
            }],
            audioRef: this.refs.audio,
            canvas: this.refs.analyzerCanvas,
            isPlaying: true,
            gainNode: {},
            creator: {},
            status: "",
        }
    }

    addActiveListener(){
        fetch('https://audio-api.kjgoodwin.me/v1/channels/' + this.state.streamID + "/listeners",{
                method: "POST",
                headers: {
                  "Authorization": GlobalStore.token
                },
            })
    }

    componentDidMount(){
        this.addActiveListener()
        var broadcastIdSplit = window.location.pathname.split("/")
        let broadcastId = broadcastIdSplit[broadcastIdSplit.length - 1];
        this.setState({streamID: broadcastId})
        this.connectToStream(this.refs.analyzerCanvas, this.refs.audio);
        GlobalStore.context.resume()
        GlobalStore.isPlaying = true;
    }

    componentWillUnmount(){

    }

    connectToStream = (canvas, audioRef) => {
        var enableRecordings = false;



        var connection = new RTCMultiConnection();

        // its mandatory in v3
        connection.enableScalableBroadcast = true;

        // each relaying-user should serve only 1 users
        connection.maxRelayLimitPerUser = 1;

        // we don't need to keep room-opened
        // scalable-broadcast.js will handle stuff itself.
        connection.autoCloseEntireSession = true;

        // by default, socket.io server is assumed to be deployed on your own URL
        //connection.socketURL = '/';

        // comment-out below line if you do not have your own socket.io server
        connection.socketURL = 'https://audio-api.kjgoodwin.me:3001/';

        connection.socketMessageEvent = 'scalable-audio-broadcast-demo';

        // document.getElementById('broadcast-id').value = connection.userid;

        // user need to connect server, so that others can reach him.
        connection.connectSocket(function (socket) {

            socket.on('logs', function (log) {
                console.log("log " + log)
            });

            // this event is emitted when a broadcast is already created.
            socket.on('join-broadcaster', function (hintsToJoinBroadcast) {
                console.log('join-broadcaster', hintsToJoinBroadcast);

                connection.session = hintsToJoinBroadcast.typeOfStreams;
                connection.sdpConstraints.mandatory = {
                    OfferToReceiveVideo: !!connection.session.video,
                    OfferToReceiveAudio: !!connection.session.audio
                };
                connection.broadcastId = hintsToJoinBroadcast.broadcastId;
                connection.join(hintsToJoinBroadcast.userid);
            });

            socket.on('rejoin-broadcast', function (broadcastId) {
                console.log('rejoin-broadcast', broadcastId);

                connection.attachStreams = [];
                socket.emit('check-broadcast-presence', broadcastId, function (isBroadcastExists) {
                    if (!isBroadcastExists) {
                        // the first person (i.e. real-broadcaster) MUST set his user-id
                        connection.userid = broadcastId;
                        // connection.userid = "gorgomish"
                    }

                    socket.emit('join-broadcast', {
                        broadcastId: broadcastId,
                        userid: connection.userid,
                        clientChannelName: connection.clientChannelName,
                        typeOfStreams: connection.session
                    });
                });
            });

            socket.on('broadcast-stopped', function (broadcastId) {
                // alert('Broadcast has been stopped.');
                // location.reload();
                console.error('broadcast-stopped', broadcastId);
                alert('This broadcast has been stopped.');
            });

            // this event is emitted when a broadcast is absent.
            socket.on('start-broadcasting', function (typeOfStreams) {
                console.log('start-broadcasting', typeOfStreams);

                // host i.e. sender should always use this!
                connection.sdpConstraints.mandatory = {
                    OfferToReceiveVideo: false,
                    OfferToReceiveAudio: false
                };
                connection.session = typeOfStreams;

                // "open" method here will capture media-stream
                // we can skip this function always; it is totally optional here.
                // we can use "connection.getUserMediaHandler" instead
                connection.mediaConstraints = {
                    video: false,
                    audio: true
                }

                var BandwidthHandler = connection.BandwidthHandler;
                connection.bandwidth = {
                    audio: 1000
                };

                connection.processSdp = function (sdp) {
                    sdp = BandwidthHandler.setApplicationSpecificBandwidth(sdp, connection.bandwidth, !!connection.session.screen);

                    sdp = BandwidthHandler.setOpusAttributes(sdp);

                    sdp = BandwidthHandler.setOpusAttributes(sdp, {
                        'stereo': 1,
                        'sprop-stereo': 1,
                        'maxaveragebitrate': connection.bandwidth.audio * 1000 * 8,
                        'maxplaybackrate': connection.bandwidth.audio * 1000 * 8,
                        'cbr': 1,
                        'useinbandfec': 1,
                        'usedtx': 1,
                        'maxptime': 3
                    });

                    return sdp;
                };


                connection.open(connection.userid);

            });

            socket.on('failed-broadcast-start', function (data) {
                console.log(data);
            });


            socket.on('broadcast-doesnt-exist', function (responseText) {
                console.log("this broadcast isnt in the database " + responseText);
            });
        });

        loadBroadcast();

        window.onbeforeunload = function () {
            // Firefox is ugly.
            document.getElementById('open-or-join').disabled = false;
        };

        //var audioPreview = this.refs.audio

        connection.onstream = function (event) {
            if (connection.isInitiator && event.type !== 'local') {
                return;
            }
            console.log("i am streaming now")
            console.log(event.stream)
            connection.isUpperUserLeft = false;
            createVisualization(event.stream, canvas, audioRef)
            //audioPreview.srcObject = event.stream;
            
            //audioPreview.play();

            //audioPreview.userid = event.userid;

            if (event.type === 'local') {
                //audioPreview.muted = true;
            }

            if (connection.isInitiator == false && event.type === 'remote') {
                // he is merely relaying the media
                connection.dontCaptureUserMedia = true;
                connection.attachStreams = [event.stream];
                
                connection.sdpConstraints.mandatory = {
                    OfferToReceiveAudio: false,
                    OfferToReceiveVideo: false
                };

                connection.getSocket(function (socket) {
                    socket.emit('can-relay-broadcast');

                    if (connection.DetectRTC.browser.name === 'Chrome') {
                        connection.getAllParticipants().forEach(function (p) {
                            if (p + '' != event.userid + '') {
                                var peer = connection.peers[p].peer;
                                peer.getLocalStreams().forEach(function (localStream) {
                                    peer.removeStream(localStream);
                                });
                                event.stream.getTracks().forEach(function (track) {
                                    peer.addTrack(track, event.stream);
                                });
                                connection.dontAttachStream = true;
                                connection.renegotiate(p);
                                connection.dontAttachStream = false;
                            }
                        });
                    }

                    if (connection.DetectRTC.browser.name === 'Firefox') {
                        // Firefox is NOT supporting removeStream method
                        // that's why using alternative hack.
                        // NOTE: Firefox seems unable to replace-tracks of the remote-media-stream
                        // need to ask all deeper nodes to rejoin
                        connection.getAllParticipants().forEach(function (p) {
                            if (p + '' != event.userid + '') {
                                connection.replaceTrack(event.stream, p);
                            }
                        });
                    }

                    // Firefox seems UN_ABLE to record remote MediaStream
                    // WebAudio solution merely records audio
                    // so recording is skipped for Firefox.
                    if (connection.DetectRTC.browser.name === 'Chrome') {
                        repeatedlyRecordStream(event.stream);
                    }
                });
            }

            // to keep room-id in cache
            localStorage.setItem(connection.socketMessageEvent, connection.sessionid);
        };

        // ask node.js server to look for a broadcast
        // if broadcast is available, simply join it. i.e. "join-broadcaster" event should be emitted.
        // if broadcast is absent, simply create it. i.e. "start-broadcasting" event should be fired.
        function loadBroadcast() {
            var broadcastIdSplit = window.location.pathname.split("/")
            let broadcastId = broadcastIdSplit[broadcastIdSplit.length - 1];
            console.log("broadcastId" + broadcastId);
            //this.setState({streamID: broadcastId})


            connection.extra.broadcastId = broadcastId;

            connection.session = {
                audio: true,
                oneway: true
            };

            connection.getSocket(function (socket) {
                socket.emit('check-broadcast-presence', broadcastId, function (isBroadcastExists) {
                    if (!isBroadcastExists) {
                        // the first person (i.e. real-broadcaster) MUST set his user-id
                        connection.userid = broadcastId;
                    }

                    console.log('check-broadcast-presence', broadcastId, isBroadcastExists);

                    var urlParams = new URLSearchParams(window.location.search);
                    let authToken = urlParams.get('auth');


                    socket.emit('join-broadcast', {
                        broadcastId: broadcastId,
                        userid: connection.userid,
                        gatewayAuth: authToken,
                        typeOfStreams: connection.session
                    });
                });
            });
        };



        connection.onstreamended = function () { };

        connection.onleave = function (event) {
            //if (event.userid !== audioPreview.userid) return;

            connection.getSocket(function (socket) {
                socket.emit('can-not-relay-broadcast');
                connection.isUpperUserLeft = true;

                if (allRecordedBlobs.length) {
                    // playing lats recorded blob
                    var lastBlob = allRecordedBlobs[allRecordedBlobs.length - 1];
                    //audioPreview.src = URL.createObjectURL(lastBlob);
                    
                    //audioPreview.play();
                    allRecordedBlobs = [];
                } else if (connection.currentRecorder) {
                    var recorder = connection.currentRecorder;
                    connection.currentRecorder = null;
                    recorder.stopRecording(function () {
                        if (!connection.isUpperUserLeft) return;

                        //audioPreview.src = URL.createObjectURL(recorder.getBlob());
                        //audioPreview.play();
                    });
                }

                if (connection.currentRecorder) {
                    connection.currentRecorder.stopRecording();
                    connection.currentRecorder = null;
                }
            });
        };

        var allRecordedBlobs = [];

        function repeatedlyRecordStream(stream) {
            if (!enableRecordings) {
                return;
            }

            /*
            connection.currentRecorder = RecordRTC(stream, {
                type: 'audio'
            });
            */

            connection.currentRecorder.startRecording();

            setTimeout(function () {
                if (connection.isUpperUserLeft || !connection.currentRecorder) {
                    return;
                }

                connection.currentRecorder.stopRecording(function () {
                    allRecordedBlobs.push(connection.currentRecorder.getBlob());

                    if (connection.isUpperUserLeft) {
                        return;
                    }

                    connection.currentRecorder = null;
                    repeatedlyRecordStream(stream);
                });
            }, 30 * 1000); // 30-seconds
        };
        function createVisualization(stream, canvas, audio){
            console.log("fucking shit")
            
            let analyser = GlobalStore.context.createAnalyser();
            let ctx = canvas.getContext('2d');
            //let audio = this.refs.audio;
            //audio.crossOrigin = "anonymous";
            console.log(stream)
            
            //this.setState({gainNode: gain});
            //let audioSrc = GlobalStore.context.createMediaStreamSource(stream);
            let audioSrc = GlobalStore.context.createMediaStreamSource(stream);
            
            
            
            audioSrc.connect(analyser);
            
            //audioSrc.connect(GlobalStore.context.destination);
           // analyser.connect(GlobalStore.context.destination);
            
            
            function renderFrame(){
                
                let freqData = new Uint8Array(analyser.frequencyBinCount)
                requestAnimationFrame(renderFrame)
                analyser.getByteFrequencyData(freqData)
                ctx.clearRect(0, 0, canvas.width, canvas.height)
                ctx.fillStyle = '#9933ff';
                let bars = 100;
                for (var i = 0; i < bars; i++) {
                    let bar_x = i * 3;
                    let bar_width = 2;
                    let bar_height = -(freqData[i] / 2);
                    ctx.fillRect(bar_x, canvas.height, bar_width, bar_height)
                }
            };
            renderFrame()
        }
    }

    playTrack = (event) => {
        event.preventDefault();
        console.log("pressed")
        if(this.state.isPlaying){
            console.log("paused")
            GlobalStore.context.suspend()
            this.setState({isPlaying: false})
        } else {
            console.log("played")
            GlobalStore.context.resume()
            this.setState({isPlaying: true})
        }
    }

    renderComment(message){
        var newMessage = JSON.parse(message)
        console.log(newMessage)
        if(newMessage.type == "status-update"){
            this.setState({status:newMessage.status.text})
        }else{
            let newMessages = this.state.comments
            newMessages.push(newMessage.message)
            this.setState({comments: newMessages})
        }
        
    }

    handleChange = event => {
        this.setState({
          [event.target.id]: event.target.value
        });
    }

    handleSubmit = async event => {
        event.preventDefault();
        var body = {body:this.state.content}
        fetch('https://audio-api.kjgoodwin.me/v1/comments/' + this.state.streamID,{
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                  "Accept": "application/json",
                  "Authorization": GlobalStore.token
                },
                body: JSON.stringify(body)
            }).then((response) => {
              return response.json()
            })
            .then((data)=>{
            });
    }
    


    render(){
        var commentList = this.state.comments.map((comment) => {
            return <Comment message={comment}></Comment>
        })
        return(
            <div className='stream-and-comments'>
                <div className='streamy'>
                    <img className='stream-profile-img' src={muaz}/>
                    <div className='stream-player-info'>
                        <div className='stream-upper'>
                        <div className='stream-title-user'>
                            <h1 className='stream-title'>{this.state.streamID}</h1>
                            <h2>Creator</h2>
                        </div>
                            <PlayButton playState={this.state.isPlaying} play={this.playTrack} />
                            <div>{this.state.status}</div>
                        </div>
                            <canvas
                            ref="analyzerCanvas"
                            id="analyzer"
                            >
                            </canvas>
                    </div>
                </div>
                <div className='comments'>
                <FormGroup controlId="content" style={{marginBottom: "8px"}}>
                    <FormControl
                    onChange={this.handleChange}
                    value={this.state.content}
                    componentClass="textarea"
                    placeholder={`Comment`}
                    size="md"
                    style={{height: 71}}
                    />
                </FormGroup>
                <button className="btn leave-feedback-btn" onClick={this.handleSubmit}>
                    Post Comment
                </button>
                <Websocket url={GlobalStore.socket}
                onMessage={this.renderComment.bind(this)}/>
                    {commentList}
                </div>
            </div>
        )
    }
}

const PlayButton = observer(({playState, play}) => 
    <div className={playState ? "track-pause-btn" : "track-play-btn"} onClick={play} />
)

class Comment extends Component{

    render(){
        return(
            <div className='comment'>
                <div className='comment-user'>
                    <div className='cropper'><img className="comment-img" src={this.props.message.creator.photoURL}></img></div>
                    <div className='comment-username'><p>{this.props.message.creator.userName}</p></div>
                </div>
                <p>{this.props.message.body}</p>
                
            </div>
        )
    }
}

export default AudioStream;
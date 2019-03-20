// recording is disabled because it is resulting for browser-crash
// if you enable below line, please also uncomment above "RecordRTC.js"

let GlobalStore = {};

let loginform = document.querySelectorAll("form")[0];
document.getElementById("email").value = "email@email.com"
document.getElementById("psw").value = "password"

loginform.addEventListener("submit", function (event) {
    event.preventDefault();
    let email = document.getElementById("email").value;
    let password = document.getElementById("psw").value;

    loginUser(email, password);
});

function loginUser(email, password) {
    let credentials = {
        email: email,
        password: password
    }


    fetch('https://audio-api.kjgoodwin.me/v1/sessions', {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
        body: JSON.stringify(credentials)
    }).then((response) => {
        if (!response.ok) {
            alert("Authentication failed")

            return
        }

        localStorage.setItem("auth", response.headers.get('Authorization'))
        return response.json()

    })
        .then((data) => {
            GlobalStore.isAuthenticated = true;
            GlobalStore.currentUser = data;

            //Hide the signin when it is don
            document.querySelectorAll(".logincontainer")[0].setAttribute("hidden", true);
            document.getElementById("channels").removeAttribute("hidden")
            getStreams();
        });

}


function getStreams() {

    fetch('https://audio-api.kjgoodwin.me/v1/channels?username=' + GlobalStore.currentUser.userName, {
        method: "GET",
        headers: {
            "Accept": "application/json",
            "Authorization": localStorage.getItem("auth")
        }
    }).then((response) => {
        if (!response.ok) {
            alert("Get channels failed")

            return
        }
        return response.json()

    })
        .then((data) => {

            console.log(data)
            GlobalStore.streams = data;

            let parent = document.getElementById("channels");
            let count = 0;
            parent.innerHTML = "";
            for (i in data) {
                let container = document.createElement("div");
                container.classList.add("stream-container");
                //container.onclick = clickStream;
                let display = document.createElement("h2");
                let channelID = document.createElement("p");
                let description = document.createElement("p");
                let button = document.createElement("button");
                button.id = count.toString();
                button.onclick = clickStream;
                button.innerText = "Start";
                display.innerText = data[i].displayName;
                channelID.innerText = "ChannelID: " + data[i].channelID;
                description.innerText = "Description: " + document.getElementById("new-description").value;
                container.appendChild(display);
                container.appendChild(channelID);
                container.appendChild(description);
                container.appendChild(button);
                parent.appendChild(container);
                count++;
            }

            document.querySelector(".new-channel").removeAttribute("hidden");
        }


        );



}

function clickStream(event) {
    //console.log(parseInt(event.target.id));
    console.log(event.target.id);
    loadBroadcast(event.target.id);
    document.getElementById("channels").setAttribute("hidden", true)
    document.querySelector(".new-channel").setAttribute("hidden", true)
    let currentStream = GlobalStore.streams[event.target.id];
    GlobalStore.currentStream = currentStream;
    let title = document.createElement("h1");
    title.innerText = "Now streaming your audio to channel: " + currentStream.channelID;
    document.querySelector("body").prepend(title);
    document.getElementById("status").removeAttribute("hidden");

}


document.getElementById("new-channel-button").onclick = addStream;


function addStream() {

    let channelID = document.getElementById("new-channelid").value;
    let displayName = document.getElementById("new-displayname").value;
    let description = document.getElementById("new-description").value;

    postStream(channelID, displayName, description);

}

function postStream(channelID, displayName, description) {
    let body = {
        channelID: channelID,
        displayName: displayName,
        description: description
    }


    fetch('https://audio-api.kjgoodwin.me/v1/channels', {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": localStorage.getItem("auth")
        },
        body: JSON.stringify(body)
    }).then((response) => {
        if (!response.ok) {
            alert("Authentication failed")
            return
        }
        return response.json()

    })
        .then((data) => {
            console.log(data)
            getStreams();
        });

}

document.getElementById("status-update").onclick = postUpdate;

function postUpdate() {

    let channelID = GlobalStore.currentStream.channelID;
    let text = document.getElementById("status-text").value;

    let body = {
        text: text,
        image: "https://blog.spoongraphics.co.uk/wp-content/uploads/2017/album-art/3.jpg"
    }

    console.log(body);


    fetch('https://audio-api.kjgoodwin.me/v1/status/' + channelID, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": localStorage.getItem("auth")
        },
        body: JSON.stringify(body)
    }).then((response) => {
        if (!response.ok) {
            alert("Update failed")
            return
        }
        return response.body;

    })
        .then((data) => {
            console.log(data);
        });

}

var enableRecordings = false;

var currentAuthUser = "";

var globalSocket;


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

    globalSocket = socket;

    socket.on('logs', function (log) {
        console.log("log " + log)
    });

    // this event is emitted when a broadcast is already created.
    socket.on('join-broadcaster', function (hintsToJoinBroadcast) {
        console.log('join-broadcaster', hintsToJoinBroadcast);
        currentAuthUser = hintsToJoinBroadcast.authUserID
        currentChannelID = hintsToJoinBroadcast.userid

        console.log(" see if channelid works " + currentChannelID)
        console.log("seeing if this works " + currentAuthUser);

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
            audio: {
                mandatory: {
                    echoCancellation: false, // disabling audio processing
                    googAutoGainControl: false,
                    googNoiseSuppression: false,
                    googHighpassFilter: false,
                    googTypingNoiseDetection: false
                    //googAudioMirroring: true
                },
                optional: []
            }
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

//loadBroadcast();

var audioPreview = document.getElementById('audio-preview');

connection.onstream = function (event) {
    if (connection.isInitiator && event.type !== 'local') {
        return;
    }

    connection.isUpperUserLeft = false;
    audioPreview.srcObject = event.stream;
    audioPreview.play();

    audioPreview.userid = event.userid;

    if (event.type === 'local') {
        audioPreview.muted = true;
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
function loadBroadcast(num) {
    //var broadcastIdSplit = window.location.pathname.split("/")
    console.log(GlobalStore.streams);
    let broadcastId = GlobalStore.streams[num].channelID;
    console.log("broadcastId" + broadcastId);


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

            //var urlParams = new URLSearchParams(window.location.search);
            let authToken = localStorage.getItem("auth");


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
    console.log(" leavingthebroadcast!!!!!!!")
    if (event.userid !== audioPreview.userid) return;

    connection.getSocket(function (socket) {
        socket.emit('can-not-relay-broadcast');
        connection.isUpperUserLeft = true;

        if (allRecordedBlobs.length) {
            // playing lats recorded blob
            var lastBlob = allRecordedBlobs[allRecordedBlobs.length - 1];
            audioPreview.src = URL.createObjectURL(lastBlob);
            audioPreview.play();
            allRecordedBlobs = [];
        } else if (connection.currentRecorder) {
            var recorder = connection.currentRecorder;
            connection.currentRecorder = null;
            recorder.stopRecording(function () {
                if (!connection.isUpperUserLeft) return;

                audioPreview.src = URL.createObjectURL(recorder.getBlob());
                audioPreview.play();
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

    connection.currentRecorder = RecordRTC(stream, {
        type: 'audio'
    });

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


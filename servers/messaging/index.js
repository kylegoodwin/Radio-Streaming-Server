const express = require("express");
const bodyParser = require('body-parser');
const app = express();



const port = process.env.PORT;
const instanceName = process.env.NAME;
const mongoPort = process.env.MONGOPORT;
const routes = require('./routes/routes.js')
const http = require('http').Server(app)

// Set up mongoose connection
const mongoose = require('mongoose');
let mongoDB = "mongodb://"+mongoPort
mongoose.connect(mongoDB);
mongoose.Promise = global.Promise;
let db = mongoose.connection;
db.on('error', console.error.bind(console, 'MongoDB connection error:'));

//check X-User header for authentication
app.use(function(req, res, next){
  let userx = req.get("X-User")
  if(userx == undefined){
    res.status = 401;
    res.send('Unauthenticated User');
  } else {
    next();
  }
});

//add JSON request body parsing middleware
app.use(express.json());
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({extended: false}));
app.use('/v1/', routes);

//TODO: add error handling function

app.listen(port, "", () => {
        //callback is executed once server is listening
        console.log(`server is listening at http://:${port}...`);
	console.log("port : " + port);
	console.log("host : " + instanceName);
});
create database if not exists db;

use db;

create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(320) not null,
    pass_hash varchar(255) not null,
    user_name varchar(255) not null, 
    first_name varchar(64) not null,
    last_name varchar(128) not null,
    photo_URL varchar(128) not null,
    UNIQUE(email),
    UNIQUE(user_name)
);

create table if not exists successful_logins (
    id int not null auto_increment primary key,
    user_id int not null,
    sign_in_time timestamp not null,
    login_ip varchar(32) not null
);

/*
create table if not exists login_attempts (
    id int not null auto_increment primary key,
    email varchar(320) not null,
    pass_hash varchar(32) not null,
    user_name varchar(255) not null, 
    first_name varchar(64) not null,
    last_name varchar(128) not null,
    photo_URL varchar(128) not null,
    UNIQUE(email),
    UNIQUE(user_name)
);
*
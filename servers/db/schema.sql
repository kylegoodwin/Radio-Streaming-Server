create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(255) not null,
    password_hash binary(60) not null,
    username varchar(255) not null,
    first_name varchar(64),
    last_name varchar(64),
    photo_url varchar(255) not null,
    index email_index (email),
    index username_index (username)
);

CREATE UNIQUE INDEX index_unique ON users(email);
CREATE UNIQUE INDEX index_unique_username ON users(username);


create table if not exists logins (
    login_key int not null auto_increment primary key,
    user_id int not null,
    login_time datetime not null,
    user_ip varchar(39) not null
);
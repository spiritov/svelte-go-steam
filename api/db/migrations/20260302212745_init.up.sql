-- example tables
create table user(
  id text not null primary key,
  alias text not null,
  avatar_url text not null,
  created_at datetime not null default current_timestamp
);

create table user_role(
  user_id text not null,
  role text not null default 'user',
  
  foreign key (user_id) references user (id)
);

-- session-related tables
create table openid_nonce(
  id integer not null primary key autoincrement,
  endpoint text not null,
  nonce_time datetime not null,
  nonce_string text not null,

  created_at datetime not null default current_timestamp,
  
  unique (endpoint, nonce_string)
);

create table session(
  id integer not null primary key autoincrement,
  user_id text not null,
  token_id text not null unique,
  
  created_at datetime not null default current_timestamp,
  foreign key (user_id) references user (id)
);

create table disallow_token(
  token_id text not null unique,

  created_at datetime not null default current_timestamp,

  foreign key (token_id) references session (token_id)
);
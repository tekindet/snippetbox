CREATE TABLE snippets(
  id SERIAL NOT NULL PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  content TEXT NOT NULL,
  created TIMESTAMP NOT NULL,
  expires TIMESTAMP NOT NULL
);

CREATE INDEX idx_snippets_created ON snippets(created);

CREATE TABLE users(
  id SERIAL NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  hashed_password CHAR(60) NOT NULL,
  created TIMESTAMP NOT NULL
);

INSERT INTO users(name,email,hashed_password,created) VALUES(
  'Indiana Jones',
  'indiana@example.com',
  '$2a$12$NuTjWXm3KKntReFwyBVHyuf/to.HEwTy.eS206TNfkGfr6HzGJSWG',
  '2022-02-21 08:45:00'
);

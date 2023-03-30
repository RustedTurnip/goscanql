-- SCHEMA
CREATE TABLE colour
(
    name  VARCHAR PRIMARY KEY,
    red   SMALLINT NOT NULL,
    green SMALLINT NOT NULL,
    blue  SMALLINT NOT NULL
);

CREATE TABLE account
(
    id            SERIAL PRIMARY KEY,
    email         VARCHAR UNIQUE NOT NULL,
    date_of_birth DATE           NOT NULL,
    created_at    TIMESTAMP      NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE TABLE pet
(
    id          SERIAL PRIMARY KEY,
    account_id  INT REFERENCES account (id) NOT NULL,
    name        VARCHAR                     NOT NULL,
    animal      VARCHAR                     NOT NULL,
    breed       VARCHAR,
    colour_name VARCHAR REFERENCES colour (name)
);

-- DATA
INSERT INTO colour (name, red, green, blue)
VALUES ('Brown', 123, 63, 0),
       ('Black', 27, 28, 23),
       ('Blonde', 240, 196, 132),
       ('Golden', 228, 172, 92);

INSERT INTO account (email, date_of_birth, created_at)
VALUES ('james.mcgill@gmail.com', '1960-11-01', '2020-05-03T16:28:02'),
       ('walter.white@gmail.com', '1958-09-07', '2020-02-22T12:56:57'),
       ('mike.ehrmantraut@gmail.com', '1942-01-01', '2020-04-14T10:14:36'),
       ('gustavo.fring@gmail.com', '1958-01-01', '2020-06-30T14:15:28');

INSERT INTO pet (name, account_id, animal, breed, colour_name)
VALUES ('Rex', (
    SELECT id
      FROM account
     WHERE email = 'james.mcgill@gmail.com'), 'Dog', 'Labrador', 'Brown'),
       ('Poppy', (
           SELECT id
             FROM account
            WHERE email = 'walter.white@gmail.com'), 'Dog', 'Alsatian', 'Black'),
       ('Daisy', (
           SELECT id
             FROM account
            WHERE email = 'walter.white@gmail.com'), 'Dog', 'Golden Retriever', 'Golden'),
       ('Mr Whiskers', (
           SELECT id
             FROM account
            WHERE email = 'mike.ehrmantraut@gmail.com'), 'Cat', 'Bengal', null),
       ('Mrs Whiskers', (
           SELECT id
             FROM account
            WHERE email = 'mike.ehrmantraut@gmail.com'), 'Cat', 'Bengal', 'Black');
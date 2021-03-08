# Ritchie Formula

## Command

````bash
rit postgres formula table
````

## Description

This formula is used to create table database.

## Dependency

This formula depends credentials:
```
CREDENTIAL_PG_HOST
CREDENTIAL_PG_DBNAME
CREDENTIAL_PG_USERNAME
CREDENTIAL_PG_PASSWORD
CREDENTIAL_PG_PORT
CREDENTIAL_PG_SSL
```
The first time that uses formula, the rit cli will ask you this credential.

You can uses the command too:
````bash
rit set credential 
? Select your provider Add a new
? Define your provider name: pg
? Define your field name: (ex.:token, secretAccessKey) host
? Select your field type: plain text
? Add more credentials fields to this provider? yes
? Define your field name: (ex.:token, secretAccessKey) dbname
? Select your field type: plain text
? Add more credentials fields to this provider? yes
? Define your field name: (ex.:token, secretAccessKey) username
? Select your field type: plain text
? Add more credentials fields to this provider? yes
? Define your field name: (ex.:token, secretAccessKey) password
? Select your field type: secret
? Add more credentials fields to this provider? yes
? Define your field name: (ex.:token, secretAccessKey) port
? Select your field type: plain text
? Add more credentials fields to this provider? yes
? Define your field name: (ex.:token, secretAccessKey) ssl
? Select your field type: plain text
? Add more credentials fields to this provider? no
? host: localhost <you database hostname>
? dbname: test <you database name>
? username: test <you database username>
? password: **** <you database password>
? port: 5432 <you database port>
? ssl: disable <you database ssl <enable|disable>>
Pg credential saved!
Check your credentials using rit list credential
````

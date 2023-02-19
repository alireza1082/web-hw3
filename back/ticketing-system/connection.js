const {Client} = require('pg')

const client = new Client({
    host: "localhost",
    user: "username",
    port: 5432,
    password: "password",
    database: "postgres"
})

module.exports = client
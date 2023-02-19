//first run bank:
//    python manage.py runserver
//then run this app:
//      nodemon ./app.js
//then you need to send a Post request to this address:
// Post    127.0.0.1:3000/
//with body input key and value=
// ['price':1234]

const express = require('express');
var querystring = require('querystring');
var http = require('http');
const { copyFileSync } = require('fs');

const app = express();

app.use(express.json());
app.use(express.urlencoded({ extended: false }))

app.get('/result/:code',(req, res)=>{
    const { code } = req.params;
    result_message = ''
    if (code === '1'){
        result_message = 'Success'
    } else if(code === '2'){
        result_message = 'Input Mismatch'
    } else if(code === '3'){
        result_message = 'Expired'
    } else if(code === '4'){
        result_message = 'No Credit'
    } else if(code === '5'){
        result_message = 'Canceled'
    }
    res.status(200).send(result_message)
    
    // TODO redirect to Website
})

app.post('/price/', (req, res) => {
    const body = req.body
    const price = parseInt(body['price'])
    console.log(price)
    

    var data = JSON.stringify({
        receipt_id:9999,
        amount:price,
        result:0,
        callback:"http://127.0.0.1:3000/result"
    });

    var options = {
        host: '127.0.0.1',
        port: 8000,
        path: '/transaction/',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(data)
        }
    };

    var bank_res
    var httpreq = http.request(options, function (response) {
        response.setEncoding('utf8');
        response.on('data', function (chunk) {
            //console.log("body: " + chunk);
            bank_res = JSON.parse(chunk)
            console.log(bank_res["id"])
        });
        // Redirects user to:  /payment/id
        response.on('end', function () {
            const addr = "http://"+options.host+":"+options.port.toString()+'/payment/'+bank_res["id"].toString()+"/"
            console.log(addr);
            res.redirect(301,addr)
        })
    });
    httpreq.write(data);
    httpreq.end();

})

app.listen(3000, () => {
    console.log('server is running on Port 3000')
});


const client = require('./connection.js')
const express = require('express');
const app = express();
var request = require('request');
// const fetch = require('node-fetch');

app.listen(3300, ()=>{
    console.log("Sever is now listening at port 3000");
})

// client.connect();

const bodyParser = require("body-parser");
app.use(bodyParser.json());

app.get('/aircraftTypes', (req, res)=>{
   client.query(`Select * from aircraft_type`, (err, result)=>{
       if(!err){
           res.send(result.rows);
       }else{ console.log(err.message) }
   });
   client.end;
})

app.get('/aircraftLayouts', (req, res)=>{
   client.query(`Select * from aircraft_layout al inner join aircraft_type at on al.type_id = at.type_id`, (err, result)=>{
       if(!err){
           res.send(result.rows);
       }else{ console.log(err.message) }
   });
   client.end;
})

app.get('/aircrafts', (req, res)=>{
   client.query(`Select * from aircraft`, (err, result)=>{
       if(!err){
           res.send(result.rows);
       }else{ console.log(err.message) }
   });
   client.end;
})

app.get('/aircraftViews', (req, res)=>{
   client.query(`Select * from aircraft_view`, (err, result)=>{
       if(!err){
           res.send(result.rows);
       }else{ console.log(err.message) }
   });
   client.end;
})

app.post('/aircraftViews/search', (req, res)=>{
   const searchModel = req.body;

   client.query(`Select * from aircraft_view where registration like '%${searchModel.registration}%' and aircraft_type like '%${searchModel.aircraft_type}%'  and type_id like '%${searchModel.type_id}%' ` , (err, result)=>{
       if(!err){
           res.send(result.rows);
       }else{ console.log(err.message) }
   });
   client.end;
})

app.get('/countries', (req, res)=>{
    client.query(`Select * from country`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })


 app.get('/cities', (req, res)=>{
    client.query(`Select * from city`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })


 app.get('/airports', (req, res)=>{
    client.query(`Select * from airport`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })

 app.get('/originDestination', (req, res)=>{
    client.query(`Select * from origin_destination`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })

 app.post('/originDestination/search', (req, res)=>{
    const searchModel = req.body;

    client.query(`Select * from origin_destination where county like '%${searchModel.county}%' and city like '%${searchModel.city}%' and airport like '%${searchModel.airport}%' and iata like '%${searchModel.iata}%'`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })


 app.get('/flights', (req, res)=>{
    client.query(`Select * from flight`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })


 app.post('/flights/search', (req, res)=>{
    const searchModel = req.body;

    client.query(`Select * from flight where flight_id like '%${searchModel.flight_id}%' and origin like '%${searchModel.origin}%' and aircraft like '%${searchModel.aircraft}%' and destination like '%${searchModel.destination}%'`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })

 app.get('/purchase', (req, res)=>{
    const token = req.header("token")

    var options = {
        uri : 'http://localhost:8000/info',
        method : 'GET',
        headers: {
            'Authorization': token,
            'Content-Type': 'application/json'
          }
    }; 

    request(options, function (error, response, body) {
        if (!error && response.statusCode == 200) {
             
            var newBody = JSON.parse(body);
            

            client.query(`Select * from purchase where corresponding_user_id = ${newBody.user_id}`, (err, result)=>{
                if(!err){
                    res.send(result.rows);
                }else{ console.log(err.message) }
            });
            client.end;
        }
        else {
            res.send('you are not login')
        }
    });


 })


 app.post('/purchase', (req, res)=>{
    const token = req.header("token")
    const purchDto = req.body;

    var options = {
        uri : 'http://localhost:8000/info',
        method : 'GET',
        headers: {
            'Authorization': token,
            'Content-Type': 'application/json'
          }
    }; 

    // var paymentOption = {
    //     uri : 'localhost:8000/transaction',
    //     method : 'POST'
    // }; 

    request(options, function (error, response, body) {
        if (!error && response.statusCode == 200) {
             
            var newBody = JSON.parse(body);
            
            client.query(`Select * from flight where flight_id = '${purchDto.flight_id}'`, (err, result)=>{
                if(!err){
               
                    if(result.rowCount == 0){
                        res.send("there is no fligh with sended serial number")
                    }else {
                            var price = 0;
                            if(purchDto.class == 'Y') {
                                price = result.rows[0].y_price;
                            } else if (purchDto.class == 'F') {
                                price = result.rows[0].f_price;
                            }else {
                                price = result.rows[0].j_price;
                            }
                    
                            let insertQuery = `insert into purchase(corresponding_user_id, title, first_name, last_name, flight_serial, offer_price, offer_class) 
                            values('${newBody.user_id}','${purchDto.title}', '${newBody.first_name}', '${newBody.last_name}', '${result.rows[0].flight_serial}', '${price}', '${purchDto.class}')`
                            
                            client.query(insertQuery, (err, result)=>{
                            if(!err){
                                 res.send('purchase create success fully')
                            //     var paymentRequest = {
                            //                         "amount":price,
                            //                         "receipt_id":21,
                            //                         "callback":"http://localhost:3300"}
                            //     request(paymentOption, {
                            //         json: true,
                            //         body:paymentRequest
                            //       }, function (error, response, body){
                            //         if (!error && response.statusCode == 200) {
                            //             var paymentBody = JSON.parse(body);
                            //             res.send(paymentBody.id);
                            //         }else{
                            //             console.log(error)
                            //             res.send("fail to connect to payment server")
                            //         }
                            //       }
                            //  )
                            }
                              else{
                                res.send("fail to create purchase")
                                 console.log(err.message) }
                            })

                    }
                }else{
                     res.send("there is no fligh with sended serial number")
                    
                }
            });
            client.end;
        }
        else {
            res.send('you are not login')
        }
    });


 })



 app.put('/purchase/afterPayment', (req, res)=>{
    const token = req.header("token")
    const purchDto = req.body;

    var options = {
        uri : 'http://localhost:8000/info',
        method : 'GET',
        headers: {
            'Authorization': token,
            'Content-Type': 'application/json'
          }
    }; 

    request(options, function (error, response, body) {
        if (!error && response.statusCode == 200) {
             
            var newBody = JSON.parse(body);
            
            var paymentStatus = purchDto.paymentStatus;
            if(paymentStatus == 1){
                res.send("payment was success full")
            }else {

            client.query(`Select * from purchase where corresponding_user_id = ${newBody.user_id} and title = '${purchDto.title}'`, (err, result)=>{
                if(!err){
               
                    if(result.rowCount == 0){
                        res.send("invalid action")
                    }else {
                            
                        let updateQuery = `delete from purchase
                        where corresponding_user_id = ${newBody.user_id}
                         and title = '${purchDto.title}'`

                        client.query(updateQuery, (err, result)=>{
                        if(!err){
                        res.send('payment faild')
                        }
                        else{ 
                            res.send('fail to cancel purchase')

                            console.log(err.message) }
                         })
                    }
                }else{
                    console.log(err);
                     res.send("there is no purchase with sended title");
                    
                }
            });
            client.end;
        }
        }
        else {
            res.send('you are not login')
        }
    });


 })

 app.get('/availableOffers', (req, res)=>{
    client.query(`Select * from available_offers`, (err, result)=>{
        if(!err){
            res.send(result.rows);
        }else{ console.log(err.message) }
    });
    client.end;
 })

client.connect();
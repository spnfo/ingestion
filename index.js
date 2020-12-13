const bodyParser = require('body-parser');
const express = require('express');
const redis = require('redis');

const app = express();
const client = redis.createClient();

client.on('error', function(error) {
	console.error(error);
});

app.use(bodyParser.json())

app.post('/intake', async (req, res) => {
	console.log(req.body);
	console.log(`${req.body.event}-${req.body.user}-pos`);

	if (Math.abs(req.body.position[0]) < 90) {
		client.set(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position), err => {
			console.log(err);
		});
	}

	setTimeout(() => {
		client.get(`${req.body.event}-${req.body.user}-chkpt`, (err, reply) => {
			res.status(200).send({checkpoint: reply});
		});
	}, 200);

});

app.listen(8000, () => console.log('Listening...'));